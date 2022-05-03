package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	dNativeGit "github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-git/go-git/v5"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type saveInfrastructureComponentsReq struct {
	Env                      string                 `json:"env"`
	InfrastructureComponents map[string]interface{} `json:"infrastructureComponents"`
}

func saveInfrastructureComponents(w http.ResponseWriter, r *http.Request) {
	var req saveInfrastructureComponentsReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logrus.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	db := r.Context().Value("store").(*store.Store)

	env, err := db.GetEnvironment(req.Env)
	if err != nil {
		logrus.Errorf("cannot get env: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*dNativeGit.RepoCache)
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(env.InfraRepo)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var stackConfig *dx.StackConfig
	stackYamlPath := filepath.Join(req.Env, "stack.yaml")
	if env.RepoPerEnv {
		stackYamlPath = "stack.yaml"
	}

	stackConfig, err = stackYaml(repo, stackYamlPath)
	if err != nil {
		if strings.Contains(err.Error(), "file not found") {
			url := stack.DefaultStackURL
			latestTag, _ := stack.LatestVersion(url)
			if latestTag != "" {
				url = url + "?tag=" + latestTag
			}

			stackConfig = &dx.StackConfig{
				Stack: dx.StackRef{
					Repository: url,
				},
			}
		} else {
			logrus.Errorf("cannot get stack yaml from repo: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	stackConfig.Config = req.InfrastructureComponents
	stackConfigBuff := bytes.NewBufferString("")
	e := yaml.NewEncoder(stackConfigBuff)
	e.SetIndent(2)
	err = e.Encode(stackConfig)
	if err != nil {
		logrus.Errorf("cannot serialize stack config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), dNativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = stack.GenerateAndWriteFiles(*stackConfig, filepath.Join(tmpPath, stackYamlPath))
	if err != nil {
		logrus.Errorf("cannot generate and write files: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	err = StageCommitAndPush(repo, token, "[Gimlet Dashboard] Updating components")
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(env.InfraRepo)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func bootstrapGitops(w http.ResponseWriter, r *http.Request) {
	bootstrapConfig := &api.GitopsBootstrapConfig{}
	err := json.NewDecoder(r.Body).Decode(&bootstrapConfig)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, gitUser, _ := tokenManager.Token()
	org := config.Github.Org

	db := r.Context().Value("store").(*store.Store)
	environment, err := db.GetEnvironment(bootstrapConfig.EnvName)
	if err != nil {
		logrus.Errorf("cannot get environment: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	environment.InfraRepo = bootstrapConfig.InfraRepo
	environment.AppsRepo = bootstrapConfig.AppsRepo
	if !strings.Contains(environment.InfraRepo, "/") {
		environment.InfraRepo = filepath.Join(org, environment.InfraRepo)
	}
	if !strings.Contains(environment.AppsRepo, "/") {
		environment.AppsRepo = filepath.Join(org, environment.AppsRepo)
	}

	environment.RepoPerEnv = bootstrapConfig.RepoPerEnv
	err = db.UpdateEnvironment(environment)
	if err != nil {
		logrus.Errorf("cannot update environment: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := ctx.Value("user").(*model.User)
	orgRepos, err := getOrgRepos(db)
	if err != nil {
		logrus.Errorf("cannot get repo list: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	isNewInfraRepo, err := AssureRepoExists(
		orgRepos,
		environment.InfraRepo,
		user.AccessToken,
		token,
		user.Login)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	isNewAppsRepo, err := AssureRepoExists(
		orgRepos,
		environment.AppsRepo,
		user.AccessToken,
		token,
		user.Login,
	)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	go updateOrgRepos(ctx)
	go updateUserRepos(config, db, user)

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*dNativeGit.RepoCache)
	infraGitopsRepoFileName, infraPublicKey, infraSecretFileName, err := BootstrapEnv(
		gitRepoCache,
		environment.Name,
		environment.InfraRepo,
		bootstrapConfig.RepoPerEnv,
		token,
		true,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	appsGitopsRepoFileName, appsPublicKey, appsSecretFileName, err := BootstrapEnv(
		gitRepoCache,
		environment.Name,
		environment.AppsRepo,
		bootstrapConfig.RepoPerEnv,
		token,
		false,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	notificationsFileName, err := BootstrapNotifications(
		gitRepoCache,
		config.GimletD.URL,
		config.GimletD.TOKEN,
		environment.Name,
		environment.AppsRepo,
		bootstrapConfig.RepoPerEnv,
		token,
		gitUser,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	guidingTexts := map[string]interface{}{
		"envName":                 bootstrapConfig.EnvName,
		"repoPerEnv":              bootstrapConfig.RepoPerEnv,
		"infraRepo":               environment.InfraRepo,
		"infraPublicKey":          infraPublicKey,
		"infraSecretFileName":     infraSecretFileName,
		"infraGitopsRepoFileName": infraGitopsRepoFileName,
		"appsRepo":                environment.AppsRepo,
		"appsPublicKey":           appsPublicKey,
		"appsSecretFileName":      appsSecretFileName,
		"appsGitopsRepoFileName":  appsGitopsRepoFileName,
		"isNewInfraRepo":          isNewInfraRepo,
		"isNewAppsRepo":           isNewAppsRepo,
		"notificationsFileName":   notificationsFileName,
	}

	guidingTextsString, err := json.Marshal(guidingTexts)
	if err != nil {
		logrus.Errorf("cannot serialize guiding texts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(guidingTextsString)
}

func BootstrapEnv(
	gitRepoCache *dNativeGit.RepoCache,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
	shouldGenerateController bool,
) (string, string, string, error) {
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot get repo: %s", err)
	}

	if repoPerEnv {
		envName = ""
	}
	gitopsRepoFileName, publicKey, secretFileName, err := gitops.GenerateManifests(
		shouldGenerateController,
		envName,
		repoPerEnv,
		tmpPath,
		true,
		true,
		fmt.Sprintf("git@github.com:%s.git", repoName),
		"main",
	)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot generate manifest: %s", err)
	}

	err = StageCommitAndPush(repo, token, "[Gimlet Dashboard] Bootstrapping")
	if err != nil {
		return "", "", "", fmt.Errorf("cannot stage commit and push: %s", err)
	}

	gitRepoCache.Invalidate(repoName)

	return gitopsRepoFileName, publicKey, secretFileName, nil
}

func BootstrapNotifications(
	gitRepoCache *dNativeGit.RepoCache,
	gimletdUrl string,
	gimletdToken string,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
	gitUser string,
) (string, error) {
	targetPath := envName
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo: %s", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	w.Pull(&git.PullOptions{
		Auth: &gitHttp.BasicAuth{
			Username: gitUser,
			Password: token,
		},
		RemoteName: "origin",
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("could not fetch: %s", err)
	}

	if repoPerEnv {
		targetPath = ""
	}
	notificationsFileName, err := gitops.GenerateManifestProviderAndAlert(
		envName,
		targetPath,
		repoPerEnv,
		tmpPath,
		fmt.Sprintf("git@github.com:%s.git", repoName),
		gimletdUrl,
		gimletdToken,
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate manifest: %s", err)
	}

	err = StageCommitAndPush(repo, token, "[Gimlet Dashboard] Bootstrapping")
	if err != nil {
		return "", fmt.Errorf("cannot stage commit and push: %s", err)
	}

	gitRepoCache.Invalidate(repoName)

	return notificationsFileName, nil
}

func AssureRepoExists(orgRepos []string,
	repoName string,
	userToken string,
	orgToken string,
	loggedInUser string,
) (bool, error) {
	if hasRepo(orgRepos, repoName) {
		return false, nil
	}

	org := ""
	parts := strings.Split(repoName, "/")
	if len(parts) == 2 {
		org = parts[0]
		repoName = parts[1]
	}

	token := orgToken
	if org == loggedInUser {
		org = "" // if the repo is not an org repo, but the logged in user's, the Github API doesn't need an org
		token = userToken
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	var (
		name     = repoName
		private  = true
		autoInit = true
	)

	r := &github.Repository{
		Name:     &name,
		Private:  &private,
		AutoInit: &autoInit,
	}
	_, _, err := client.Repositories.Create(context.Background(), org, r)

	return true, err
}

func StageCommitAndPush(repo *git.Repository, token string, msg string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	_, err = nativeGit.Commit(repo, msg)
	if err != nil {
		return err
	}

	err = nativeGit.PushWithToken(repo, token)
	if err != nil {
		return err
	}

	return nil
}
