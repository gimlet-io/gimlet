package server

import (
	"bytes"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-chi/chi"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
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
		logrus.Errorf("cannot decode req: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	db := r.Context().Value("store").(*store.Store)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	config := ctx.Value("config").(*config.Config)
	user := ctx.Value("user").(*model.User)
	goScm := genericScm.NewGoScmHelper(config, nil)

	env, err := db.GetEnvironment(req.Env)
	if err != nil {
		logrus.Errorf("cannot get env: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
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

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sourceBranch, err := GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-stack-change-%s", env.Name), 4)
	if err != nil {
		logrus.Errorf("cannot generate branch name: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = nativeGit.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
	if err != nil {
		logrus.Errorf("cannot checkout branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), nativeGit.Dir_RWX_RX_R)
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

	err = StageCommitAndPush(repo, tmpPath, token, "[Gimlet Dashboard] Updating components")
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	createdPR, _, err := goScm.CreatePR(token, env.InfraRepo, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet Dashboard] `%s` infrastructure components change", env.Name),
		fmt.Sprintf("@%s is editing the infrastructure components on `%s`", user.Login, env.Name))
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(env.InfraRepo)

	response := map[string]interface{}{
		"envName": env.Name,
		"createdPr": &api.PR{
			Sha:     createdPR.Sha,
			Link:    createdPR.Link,
			Title:   createdPR.Title,
			Source:  createdPR.Source,
			Number:  createdPR.Number,
			Author:  createdPR.Author.Login,
			Created: int(createdPR.Created.Unix()),
			Updated: int(createdPR.Updated.Unix()),
		},
		"stackConfig": stackConfig,
	}

	responseString, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("cannot serialize stack config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseString))
}

func bootstrapGitops(w http.ResponseWriter, r *http.Request) {
	bootstrapConfig := &api.GitopsBootstrapConfig{}
	err := json.NewDecoder(r.Body).Decode(&bootstrapConfig)
	if err != nil {
		logrus.Errorf("cannot decode bootstrap config: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	gitServiceImpl := ctx.Value("gitService").(customScm.CustomGitService)
	gitToken, gitUser, _ := tokenManager.Token()
	org := config.Org()

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
	environment.KustomizationPerApp = bootstrapConfig.KusomizationPerApp
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

	_, err = AssureRepoExists(
		orgRepos,
		environment.InfraRepo,
		user.AccessToken,
		gitToken,
		user.Login,
		gitServiceImpl,
	)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, err = AssureRepoExists(
		orgRepos,
		environment.AppsRepo,
		user.AccessToken,
		gitToken,
		user.Login,
		gitServiceImpl,
	)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	go updateOrgRepos(ctx)
	go updateUserRepos(config, db, user)

	scmURL := config.ScmURL()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	_, _, err = BootstrapEnv(
		gitRepoCache,
		gitServiceImpl,
		environment.Name,
		environment.InfraRepo,
		bootstrapConfig.RepoPerEnv,
		gitToken,
		true,
		true,
		false,
		false,
		scmURL,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _, err = BootstrapEnv(
		gitRepoCache,
		gitServiceImpl,
		environment.Name,
		environment.AppsRepo,
		bootstrapConfig.RepoPerEnv,
		gitToken,
		false,
		false,
		environment.KustomizationPerApp,
		true,
		scmURL,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fluxUser := &model.User{
		Login:  fmt.Sprintf(fluxPattern, bootstrapConfig.EnvName),
		Secret: base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
	}

	err = db.CreateUser(fluxUser)
	if err != nil {
		logrus.Errorf("cannot create user %s: %s", fluxUser.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	token := token.New(token.UserToken, fluxUser.Login)
	tokenStr, err := token.Sign(fluxUser.Secret)
	if err != nil {
		logrus.Errorf("couldn't create user token %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	_, err = BootstrapNotifications(
		gitRepoCache,
		config.Host,
		tokenStr,
		environment.Name,
		environment.AppsRepo,
		bootstrapConfig.RepoPerEnv,
		gitToken,
		gitUser,
		scmURL,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	guidingTexts := map[string]interface{}{
		"envName": bootstrapConfig.EnvName,
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
	gitRepoCache *nativeGit.RepoCache,
	gitServiceImpl customScm.CustomGitService,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
	shouldGenerateController bool,
	shouldGenerateDependencies bool,
	kustomizationPerApp bool,
	deployKeyCanWrite bool,
	scmURL string,
) (string, string, error) {
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		if strings.Contains(err.Error(), "remote repository is empty") {
			repo, tmpPath, err = initRepo(scmURL, repoName)
			defer os.RemoveAll(tmpPath)
			if err != nil {
				return "", "", fmt.Errorf("cannot init empty repo: %s", err)
			}
		} else {
			return "", "", fmt.Errorf("cannot get repo: %s", err)
		}
	}

	if repoPerEnv {
		envName = ""
	}
	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		return "", "", fmt.Errorf("cannot get head branch: %s", err)
	}

	scmHost := strings.Split(scmURL, "://")[1]
	gitopsRepoFileName, publicKey, secretFileName, err := gitops.GenerateManifests(
		shouldGenerateController,
		shouldGenerateDependencies,
		kustomizationPerApp,
		envName,
		repoPerEnv,
		tmpPath,
		true,
		true,
		fmt.Sprintf("git@%s:%s.git", scmHost, repoName),
		headBranch,
	)
	if err != nil {
		return "", "", fmt.Errorf("cannot generate manifest: %s", err)
	}

	err = StageCommitAndPush(repo, tmpPath, token, "[Gimlet Dashboard] Bootstrapping")
	if err != nil {
		return "", "", fmt.Errorf("cannot stage commit and push: %s", err)
	}

	owner, repository := ParseRepo(repoName)
	err = gitServiceImpl.AddDeployKeyToRepo(
		owner,
		repository,
		token,
		"flux",
		publicKey,
		deployKeyCanWrite,
	)
	if err != nil {
		return "", "", fmt.Errorf("cannot add deploy key to repo: %s", err)
	}

	gitRepoCache.Invalidate(repoName)

	return gitopsRepoFileName, secretFileName, nil
}

func initRepo(scmURL string, repoName string) (*git.Repository, string, error) {
	tmpPath, _ := ioutil.TempDir("", "gitops-")
	repo, err := git.PlainInit(tmpPath, false)
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	w, err := repo.Worktree()
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	err = nativeGit.StageFile(w, "", "README.md")
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	_, err = nativeGit.Commit(repo, "Init")
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	_, err = repo.CreateRemote(&gitConfig.RemoteConfig{
		Name: "origin",
		URLs: []string{fmt.Sprintf("%s/%s.git", scmURL, repoName)},
	})
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}

	return repo, tmpPath, nil
}

func BootstrapNotifications(
	gitRepoCache *nativeGit.RepoCache,
	gimletdUrl string,
	gimletdToken string,
	envName string,
	repoName string,
	repoPerEnv bool,
	token string,
	gitUser string,
	scmURL string,
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
	scmHost := strings.Split(scmURL, "://")[1]
	notificationsFileName, err := gitops.GenerateManifestProviderAndAlert(
		envName,
		targetPath,
		repoPerEnv,
		tmpPath,
		fmt.Sprintf("git@%s:%s.git", scmHost, repoName),
		gimletdUrl,
		gimletdToken,
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate manifest: %s", err)
	}

	err = StageCommitAndPush(repo, tmpPath, token, "[Gimlet Dashboard] Bootstrapping")
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
	gitServiceImpl customScm.CustomGitService,
) (bool, error) {
	if hasRepo(orgRepos, repoName) {
		return false, nil
	}

	owner := ""
	parts := strings.Split(repoName, "/")
	if len(parts) == 2 {
		owner = parts[0]
		repoName = parts[1]
	}

	err := gitServiceImpl.CreateRepository(owner, repoName, loggedInUser, orgToken, userToken)
	return true, err
}

func StageCommitAndPush(repo *git.Repository, tmpPath string, token string, msg string) error {
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

	// Temporarily staging deleted files to git with a simple CLI command until the
	// following issue is not solved:
	// https://github.com/go-git/go-git/issues/223
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpPath
	err = cmd.Run()
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

func installAgent(w http.ResponseWriter, r *http.Request) {
	envName := chi.URLParam(r, "env")

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	db := r.Context().Value("store").(*store.Store)

	env, err := db.GetEnvironment(envName)
	if err != nil {
		logrus.Errorf("cannot get environment: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(env.InfraRepo)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stackYamlPath := filepath.Join(env.Name, "stack.yaml")
	if env.RepoPerEnv {
		stackYamlPath = "stack.yaml"
	}

	stackConfig, err := stackYaml(repo, stackYamlPath)
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
				Config: map[string]interface{}{},
			}
		} else {
			logrus.Errorf("cannot get stack yaml from repo: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	agentKey := r.Context().Value("agentJWT").(string)
	stackConfig.Config["gimletAgent"] = map[string]interface{}{
		"enabled":          true,
		"environment":      env.Name,
		"agentKey":         agentKey,
		"dashboardAddress": config.Host,
	}

	stackConfigBuff := bytes.NewBufferString("")
	e := yaml.NewEncoder(stackConfigBuff)
	e.SetIndent(2)
	err = e.Encode(stackConfig)
	if err != nil {
		logrus.Errorf("cannot serialize stack config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sourceBranch, err := GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-stack-change-%s", env.Name), 4)
	if err != nil {
		logrus.Errorf("cannot generate branch name: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = nativeGit.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
	if err != nil {
		logrus.Errorf("cannot checkout branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), nativeGit.Dir_RWX_RX_R)
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
	err = StageCommitAndPush(repo, tmpPath, token, "[Gimlet Dashboard] Updating components")
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	goScm := genericScm.NewGoScmHelper(config, nil)
	user := ctx.Value("user").(*model.User)
	createdPR, _, err := goScm.CreatePR(token, env.InfraRepo, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet Dashboard] `%s` infrastructure components change", env.Name),
		fmt.Sprintf("@%s is editing the infrastructure components on `%s`", user.Login, env.Name))
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(env.InfraRepo)

	response := map[string]interface{}{
		"envName": env.Name,
		"createdPr": &api.PR{
			Sha:     createdPR.Sha,
			Link:    createdPR.Link,
			Title:   createdPR.Title,
			Source:  createdPR.Source,
			Number:  createdPR.Number,
			Author:  createdPR.Author.Login,
			Created: int(createdPR.Created.Unix()),
			Updated: int(createdPR.Updated.Unix()),
		},
		"stackConfig": stackConfig,
	}

	responseString, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("cannot serialize stack config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseString))
}

func ParseRepo(repoName string) (string, string) {
	owner := strings.Split(repoName, "/")[0]
	repo := strings.Split(repoName, "/")[1]
	return owner, repo
}
