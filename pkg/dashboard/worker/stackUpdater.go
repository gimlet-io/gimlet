package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	update "github.com/gimlet-io/gimlet-cli/pkg/commands/stack"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type stackUpdater struct {
	store         *store.Store
	dynamicConfig *dynamicconfig.DynamicConfig
	tokenManager  customScm.NonImpersonatedTokenManager
	repoCache     *nativeGit.RepoCache
}

func NewStackUpdater(
	store *store.Store,
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *nativeGit.RepoCache,
) *stackUpdater {
	return &stackUpdater{
		store:         store,
		dynamicConfig: dynamicConfig,
		tokenManager:  tokenManager,
		repoCache:     repoCache,
	}
}

func (u *stackUpdater) Run() {
	for {
		token, _, _ := u.tokenManager.Token()
		envsFromDB, err := u.store.GetEnvironments()
		if err != nil {
			logrus.Errorf("cannot get envs from db: %s", err)
		}

		for _, env := range envsFromDB {
			if env.BuiltIn {
				continue
			}

			err := updateGimletStack(
				u.dynamicConfig,
				u.repoCache,
				env.Name,
				env.InfraRepo,
				env.RepoPerEnv,
				token,
			)
			if err != nil {
				logrus.Errorf("cannot update stack: %s", err)
			}
		}

		logrus.Info("stack update process completed")
		time.Sleep(24 * time.Hour)
	}
}

func updateGimletStack(
	dynamicConfig *dynamicconfig.DynamicConfig,
	repoCache *nativeGit.RepoCache,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
) error {
	logrus.Infof("evaluating %s for stack update", repoName)
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, nil)
	prList, err := goScmHelper.ListOpenPRs(token, repoName)
	if err != nil {
		return fmt.Errorf("cannot list pull requests: %s", err)
	}

	for _, pullRequest := range prList {
		if strings.HasPrefix(pullRequest.Source, fmt.Sprintf("gimlet-stack-update-%s", envName)) {
			return nil
		}
	}

	repo, tmpPath, err := repoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot get repo: %s", err)
	}

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		return fmt.Errorf("cannot get head branch: %s", err)
	}

	stackPath := filepath.Join(envName, "stack.yaml")
	if repoPerEnv {
		stackPath = "stack.yaml"
	}

	yamlString, err := nativeGit.RemoteContentOnBranchWithoutCheckout(repo, headBranch, stackPath)
	if err != nil {
		return err
	}

	var stackConfig dx.StackConfig
	err = yaml.Unmarshal([]byte(yamlString), &stackConfig)
	if err != nil {
		return err
	}

	currentTagString := stack.CurrentVersion(stackConfig.Stack.Repository)
	versionsSince, err := stack.VersionsSince(stackConfig.Stack.Repository, currentTagString)
	if err != nil {
		logrus.Infof("Cannot check for updates: %s", err.Error())
	}

	if len(versionsSince) == 0 {
		return nil
	}

	sourceBranch, err := server.GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-stack-update-%s", envName), 4)
	if err != nil {
		return fmt.Errorf("cannot generate branch name: %s", err)
	}

	err = nativeGit.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
	if err != nil {
		return fmt.Errorf("cannot checkout branch: %s", err)
	}

	latestTag, err := stack.LatestVersion(stackConfig.Stack.Repository)
	if err != nil {
		fmt.Printf("%v  cannot find latest version\n", emoji.CrossMark)
	}

	stackConfig.Stack.Repository = stack.RepoUrlWithoutVersion(stackConfig.Stack.Repository) + "?tag=" + latestTag
	err = stack.WriteStackConfig(stackConfig, filepath.Join(tmpPath, stackPath))
	if err != nil {
		return fmt.Errorf("cannot write stack file %s", err)
	}

	err = stack.GenerateAndWriteFiles(stackConfig, filepath.Join(tmpPath, stackPath))
	if err != nil {
		return fmt.Errorf("could not generate and write files: %s", err.Error())
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return fmt.Errorf("cannot get git state: %s", err)
	}
	if empty {
		return nil
	}

	err = server.StageCommitAndPush(repo, tmpPath, token, "[Gimlet] Gimlet stack update")
	if err != nil {
		return fmt.Errorf("cannot stage, commit and push: %s", err)
	}

	changeLog, err := changeLog(stackConfig, versionsSince)
	if err != nil {
		return err
	}

	_, _, err = goScmHelper.CreatePR(token, repoName, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet] Stack update ➡️ `%s` from %s to %s", envName, currentTagString, latestTag),
		changeLog)
	if err != nil {
		return fmt.Errorf("cannot create pull request: %s", err)
	}
	logrus.Infof("pull request created for %s with stack update", repoName)

	return nil
}

func changeLog(stackConfig dx.StackConfig, versions []string) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v  Change log:\n\n", emoji.Books))
	for _, version := range versions {
		sb.WriteString(fmt.Sprintf("   - %s \n", version))

		repoUrl := stackConfig.Stack.Repository
		repoUrl = stack.RepoUrlWithoutVersion(repoUrl)
		repoUrl = repoUrl + "?tag=" + version

		stackDefinitionYaml, err := stack.StackDefinitionFromRepo(repoUrl)
		if err != nil {
			return "", fmt.Errorf("cannot get stack definition: %s", err.Error())
		}
		var stackDefinition update.StackDefinition
		err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
		if err != nil {
			return "", fmt.Errorf("cannot parse stack definition: %s", err.Error())
		}

		if stackDefinition.ChangLog != "" {
			changeLog := markdown.Render(stackDefinition.ChangLog, 80, 6)
			sb.WriteString(fmt.Sprintf("%s\n", changeLog))
		}
	}

	return sb.String(), nil
}
