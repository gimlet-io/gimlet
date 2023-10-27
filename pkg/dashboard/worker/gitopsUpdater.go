package worker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/sirupsen/logrus"
)

const (
	shouldGenerateController   = true
	shouldGenerateDependencies = true
)

type GitopsUpdater struct {
	store              *store.Store
	dynamicConfig      *dynamicconfig.DynamicConfig
	tokenManager       customScm.NonImpersonatedTokenManager
	repoCache          *nativeGit.RepoCache
	gitopsUpdatePrList *map[string]interface{}
}

func NewGitopsUpdater(
	store *store.Store,
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *nativeGit.RepoCache,
	gitopsUpdatePrList *map[string]interface{},
) *GitopsUpdater {
	return &GitopsUpdater{
		store:              store,
		dynamicConfig:      dynamicConfig,
		tokenManager:       tokenManager,
		repoCache:          repoCache,
		gitopsUpdatePrList: gitopsUpdatePrList,
	}
}

func (u *GitopsUpdater) Run() {
	for {
		token, _, _ := u.tokenManager.Token()
		scmUrl := u.dynamicConfig.ScmURL()
		envsFromDB, err := u.store.GetEnvironments()
		if err != nil {
			logrus.Errorf("cannot get envs from db: %s", err)
		}

		for _, env := range envsFromDB {
			if env.BuiltIn {
				continue
			}

			gitopsUpdatePRs := []*api.PR{}
			infraPr, err := updateGitopsManifests(
				u.dynamicConfig,
				u.repoCache,
				env.Name,
				env.InfraRepo,
				env.RepoPerEnv,
				token,
				shouldGenerateController,
				shouldGenerateDependencies,
				env.KustomizationPerApp,
				scmUrl,
			)
			if err != nil {
				logrus.Errorf("cannot update infra repo manifests: %s", err)
			}

			if infraPr != nil {
				gitopsUpdatePRs = append(gitopsUpdatePRs, infraPr)
			}

			appsPr, err := updateGitopsManifests(
				u.dynamicConfig,
				u.repoCache,
				env.Name,
				env.AppsRepo,
				env.RepoPerEnv,
				token,
				!shouldGenerateController,
				!shouldGenerateDependencies,
				env.KustomizationPerApp,
				scmUrl,
			)
			if err != nil {
				logrus.Errorf("cannot update apps repo manifests: %s", err)
			}

			if appsPr != nil {
				gitopsUpdatePRs = append(gitopsUpdatePRs, appsPr)
			}

			(*u.gitopsUpdatePrList)[env.Name] = gitopsUpdatePRs
		}

		logrus.Info("gitops update process completed")
		time.Sleep(24 * time.Hour)
	}
}

func updateGitopsManifests(
	dynamicConfig *dynamicconfig.DynamicConfig,
	repoCache *nativeGit.RepoCache,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
	shouldGenerateController bool,
	shouldGenerateDependencies bool,
	kustomizationPerApp bool,
	scmURL string,
) (*api.PR, error) {
	logrus.Infof("evaluating %s for gitops update", repoName)
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, nil)
	prList, err := goScmHelper.ListOpenPRs(token, repoName)
	if err != nil {
		return nil, fmt.Errorf("cannot list pull requests: %s", err)
	}

	for _, pullRequest := range prList {
		if strings.HasPrefix(pullRequest.Source, fmt.Sprintf("gimlet-gitops-update-%s", envName)) {
			return &api.PR{
				Sha:    pullRequest.Sha,
				Link:   pullRequest.Link,
				Title:  pullRequest.Title,
				Number: pullRequest.Number,
			}, nil
		}
	}

	repo, tmpPath, err := repoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get repo: %s", err)
	}

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get head branch: %s", err)
	}

	sourceBranch, err := server.GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-gitops-update-%s", envName), 4)
	if err != nil {
		return nil, fmt.Errorf("cannot generate branch name: %s", err)
	}

	err = nativeGit.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
	if err != nil {
		return nil, fmt.Errorf("cannot checkout branch: %s", err)
	}

	env := envName
	if repoPerEnv {
		env = ""
	}

	scmHost := strings.Split(scmURL, "://")[1]
	_, _, _, err = gitops.GenerateManifests(gitops.ManifestOpts{
		ShouldGenerateController:           shouldGenerateController,
		ShouldGenerateDependencies:         shouldGenerateDependencies,
		KustomizationPerApp:                kustomizationPerApp,
		Env:                                env,
		SingleEnv:                          repoPerEnv,
		GitopsRepoPath:                     tmpPath,
		ShouldGenerateKustomizationAndRepo: true,
		ShouldGenerateDeployKey:            false,
		GitopsRepoUrl:                      fmt.Sprintf("git@%s:%s.git", scmHost, repoName),
		Branch:                             headBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate manifest: %s", err)
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return nil, fmt.Errorf("cannot get git state: %s", err)
	}
	if empty {
		return nil, nil
	}

	err = server.StageCommitAndPush(repo, tmpPath, token, "[Gimlet] Gitops manifests update")
	if err != nil {
		return nil, fmt.Errorf("cannot stage, commit and push: %s", err)
	}

	createdPr, _, err := goScmHelper.CreatePR(token, repoName, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet] `%s` gitops manifests update for %s", repoName, envName),
		"This is an automated Pull Request that updates the Gitops manifests.")
	if err != nil {
		return nil, fmt.Errorf("cannot create pull request: %s", err)
	}
	logrus.Infof("pull request created for %s with gitops update", repoName)

	return &api.PR{
		Sha:    createdPr.Sha,
		Link:   createdPr.Link,
		Title:  createdPr.Title,
		Number: createdPr.Number,
	}, nil
}
