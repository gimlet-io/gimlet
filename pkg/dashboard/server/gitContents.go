package server

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	helper "github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-chi/chi/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func branches(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	repoName := fmt.Sprintf("%s/%s", owner, name)

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	var refIter storer.ReferenceIter
	branches := []string{}
	gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
		refIter, _ = repo.References()
		return nil
	})
	refIter.ForEach(func(r *plumbing.Reference) error {
		if r.Name().IsRemote() {
			branch := r.Name().Short()
			branches = append(branches, strings.TrimPrefix(branch, "origin/"))
		}
		return nil
	})

	branchesString, err := json.Marshal(branches)
	if err != nil {
		logrus.Errorf("cannot serialize branches: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(branchesString)
}

func getMetas(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "name")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	var err error
	var headBranch string
	var files map[string]string
	err = gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
		var innerErr error

		headBranch, innerErr = nativeGit.HeadBranch(repo)
		if innerErr != nil {
			return innerErr
		}

		files, innerErr = helper.RemoteFolderOnBranchWithoutCheckout(repo, "", ".gimlet")
		return innerErr
	})
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			logrus.Errorf("cannot list files in .gimlet/: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	fileInfos := []fileInfo{}
	for fileName, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			AppName:  envConfig.App,
			EnvName:  envConfig.Env,
			FileName: fileName,
			Branch:   headBranch,
		})
	}

	gitRepoM := gitRepoMetas{
		FileInfos: fileInfos,
	}

	gitRepoMString, err := json.Marshal(gitRepoM)
	if err != nil {
		logrus.Errorf("cannot serialize repo meta: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(gitRepoMString)
}

func configChangePullRequestsPerConfig(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	env := chi.URLParam(r, "env")
	config := chi.URLParam(r, "config")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)

	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	prList, err := goScm.ListOpenPRs(token, repoPath)
	if err != nil {
		logrus.Errorf("cannot list pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	pullRequestsPerEnvConfig := []*api.PR{}
	for _, pullRequest := range prList {
		if !strings.HasPrefix(pullRequest.Source, "gimlet-config-change") {
			continue
		}

		if !strings.Contains(pullRequest.Source, env) {
			continue
		}

		if !strings.Contains(pullRequest.Title, fmt.Sprintf("`%s`", config)) {
			continue
		}

		pullRequestsPerEnvConfig = append(pullRequestsPerEnvConfig, &api.PR{
			Branch:  pullRequest.Source,
			Sha:     pullRequest.Sha,
			Link:    pullRequest.Link,
			Title:   pullRequest.Title,
			Number:  pullRequest.Number,
			Author:  pullRequest.Author.Login,
			Created: int(pullRequest.Created.Unix()),
			Updated: int(pullRequest.Updated.Unix()),
		})
	}

	pullRequestsString, err := json.Marshal(pullRequestsPerEnvConfig)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(pullRequestsString)
}

func getPullRequests(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)

	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	prList, err := goScm.ListOpenPRs(token, repoPath)
	if err != nil {
		logrus.Errorf("cannot list pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	pullRequests := []*api.PR{}
	for _, pullRequest := range prList {
		pullRequests = append(pullRequests, &api.PR{
			Branch:  pullRequest.Source,
			Sha:     pullRequest.Sha,
			Link:    pullRequest.Link,
			Title:   pullRequest.Title,
			Number:  pullRequest.Number,
			Author:  pullRequest.Author.Login,
			Created: int(pullRequest.Created.Unix()),
			Updated: int(pullRequest.Updated.Unix()),
		})
	}

	pullRequestsString, err := json.Marshal(pullRequests)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(pullRequestsString)
}

func getChartUpdatePullRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chartUpdatePullRequests := ctx.Value("chartUpdatePullRequests").(*map[string]interface{})

	pullRequestsString, err := json.Marshal(chartUpdatePullRequests)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(pullRequestsString)
}

func getGitopsUpdatePullRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	env := chi.URLParam(r, "env")
	gitopsUpdatePullRequests := ctx.Value("gitopsUpdatePullRequests").(*map[string]interface{})

	gitopsUpdatePullRequestsPerEnv := []*api.PR{}
	if (*gitopsUpdatePullRequests)[env] != nil {
		gitopsUpdatePullRequestsPerEnv = (*gitopsUpdatePullRequests)[env].([]*api.PR)
	}

	gitopsUpdatePullRequestsString, err := json.Marshal(gitopsUpdatePullRequestsPerEnv)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(gitopsUpdatePullRequestsString)
}

func getPullRequestsFromInfraRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	envParam := r.URL.Query().Get("env")
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	db := r.Context().Value("store").(*store.Store)
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get all environments from database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	prListCreatedByGimlet := []*api.PR{}
	for _, env := range envsFromDB {
		owner, _ := scm.Split(env.InfraRepo)
		if owner == "builtin" || env.Name != envParam {
			continue
		}

		prList, err := goScm.ListOpenPRs(token, env.InfraRepo)
		if err != nil {
			if !strings.Contains(err.Error(), "Not Found") {
				logrus.Errorf("cannot list pull requests: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		for _, pullRequest := range prList {
			if strings.HasPrefix(pullRequest.Source, "gimlet-stack") && strings.Contains(pullRequest.Source, env.Name) {
				prListCreatedByGimlet = append(prListCreatedByGimlet, &api.PR{
					Sha:     pullRequest.Sha,
					Link:    pullRequest.Link,
					Title:   pullRequest.Title,
					Branch:  pullRequest.Source,
					Number:  pullRequest.Number,
					Author:  pullRequest.Author.Login,
					Created: int(pullRequest.Created.Unix()),
					Updated: int(pullRequest.Updated.Unix()),
				})
			}
		}
	}

	PullRequestsString, err := json.Marshal(prListCreatedByGimlet)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(PullRequestsString)
}

type fileInfo struct {
	EnvName  string `json:"envName"`
	AppName  string `json:"appName"`
	FileName string `json:"fileName"`
	Branch   string `json:"branch"`
}

type gitRepoMetas struct {
	FileInfos         []fileInfo `json:"fileInfos"`
	PullRequestPolicy bool       `json:"pullRequestPolicy"`
}

// envConfig fetches all environment configs from source control for a repo
func envConfigs(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "name")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	var files map[string]string
	var err error
	err = gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
		var innerErr error
		files, innerErr = helper.RemoteFolderOnBranchWithoutCheckout(repo, "", ".gimlet")
		return innerErr
	})
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			return
		} else {
			logrus.Errorf("cannot list files in .gimlet/: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	envConfigs := []dx.Manifest{}
	for _, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		envConfigs = append(envConfigs, envConfig)
	}

	configsPerEnv := map[string][]dx.Manifest{}
	for _, config := range envConfigs {
		if configsPerEnv[config.Env] == nil {
			configsPerEnv[config.Env] = []dx.Manifest{}
		}

		configsPerEnv[config.Env] = append(configsPerEnv[config.Env], config)
	}

	configsPerEnvJson, err := json.Marshal(configsPerEnv)
	if err != nil {
		logrus.Errorf("cannot convert envconfigs to json: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(configsPerEnvJson))
}

type DeploymentTemplate struct {
	Reference config.DefaultChart `json:"reference"`
	Schema    interface{}         `json:"schema"`
	UISchema  interface{}         `json:"uiSchema"`
}

func saveEnvConfig(w http.ResponseWriter, r *http.Request) {
	var envConfigData dx.Manifest
	err := json.NewDecoder(r.Body).Decode(&envConfigData)
	if err != nil {
		logrus.Errorf("cannot decode env config data: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)
	env := envConfigData.Env

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	user := ctx.Value("user").(*model.User)
	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)

	repo, tmpPath, err := gitRepoCache.InstanceForWrite(fmt.Sprintf("%s/%s", owner, repoName))
	defer os.RemoveAll(tmpPath)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	headBranch, err := helper.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	existingEnvConfigs, err := existingEnvConfigs(repo, headBranch)
	if err != nil {
		logrus.Errorf(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	envConfigFileName := envConfigPath(env, envConfigData.App, existingEnvConfigs)
	envConfigFilePath := fmt.Sprintf(".gimlet/%s", envConfigFileName)

	createCase := false // indicates if we need to create the file, we update existing manifests on the default path
	_, _, err = goScm.Content(token, repoPath, url.QueryEscape(envConfigFilePath), headBranch)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			createCase = true
		} else {
			logrus.Errorf("cannot fetch envConfig from github: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			w.Write([]byte("{}"))
			return
		}
	}

	if !createCase && existingEnvConfigs[envConfigFileName].App != envConfigData.App { // we do not allow updating application names
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// marshall and ident the manifest
	var toSaveBuffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&toSaveBuffer)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&envConfigData)
	if err != nil {
		logrus.Errorf("cannot marshal manifest: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dao := ctx.Value("store").(*store.Store)
	pullRequestPolicy, err := dao.RepoHasPullRequestPolicy(repoPath)
	if err != nil {
		logrus.Errorf("cannot get pull request policy for %s: %s", repoPath, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var sourceBranch string
	if pullRequestPolicy {
		// generate branch name to write changes on
		sourceBranch, err = GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-config-change-%s", env), 4)
		if err != nil {
			logrus.Errorf("cannot generate branch name: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = helper.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
		if err != nil {
			logrus.Errorf("cannot checkout branch: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	message := fmt.Sprintf("[Gimlet] Updating %s gimlet manifest for the %s env", envConfigData.App, env)
	if createCase {
		message = fmt.Sprintf("[Gimlet] Creating %s gimlet manifest for the %s env", envConfigData.App, env)
	}

	_ = os.MkdirAll(filepath.Join(tmpPath, ".gimlet"), nativeGit.Dir_RWX_RX_R)
	err = os.WriteFile(filepath.Join(tmpPath, envConfigFilePath), toSaveBuffer.Bytes(), nativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	hash, err := stageCommitAndPushWithHash(repo, tmpPath, token, message)
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !pullRequestPolicy {
		response := map[string]interface{}{
			"link": fmt.Sprintf("%s/%s/commit/%s", dynamicConfig.ScmURL(), repoPath, hash),
		}

		responseJson, err := json.Marshal(response)
		if err != nil {
			logrus.Errorf("cannot marshal response: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseJson)
		return
	}

	createdPR, _, err := goScm.CreatePR(token, repoPath, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet] `%s` ➡️ `%s` deployment configuration change", envConfigData.App, env),
		fmt.Sprintf("@%s is editing the `%s` deployment configuration for the `%s` environment.", user.Login, envConfigData.App, env))
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"link": createdPR.Link,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("cannot marshal response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(repoPath)
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func deleteEnvConfig(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)
	env := chi.URLParam(r, "env")
	configName := chi.URLParam(r, "config")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	user := ctx.Value("user").(*model.User)
	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)

	repo, tmpPath, err := gitRepoCache.InstanceForWrite(fmt.Sprintf("%s/%s", owner, repoName))
	defer os.RemoveAll(tmpPath)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	headBranch, err := helper.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	existingEnvConfigs, err := existingEnvConfigs(repo, headBranch)
	if err != nil {
		logrus.Errorf(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	envConfigFileName := envConfigPath(env, configName, existingEnvConfigs)
	envConfigFilePath := fmt.Sprintf(".gimlet/%s", envConfigFileName)
	_, _, err = goScm.Content(token, repoPath, envConfigFilePath, headBranch)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			logrus.Errorf("config file not found: %s", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			w.Write([]byte("{}"))
			return
		} else {
			logrus.Errorf("cannot fetch envConfig from github: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			w.Write([]byte("{}"))
			return
		}
	}

	dao := ctx.Value("store").(*store.Store)
	pullRequestPolicy, err := dao.RepoHasPullRequestPolicy(repoPath)
	if err != nil {
		logrus.Errorf("cannot get pull request policy for %s: %s", repoPath, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var sourceBranch string
	if pullRequestPolicy {
		sourceBranch, err = GenerateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-config-change-%s", env), 4)
		if err != nil {
			logrus.Errorf("cannot generate branch name: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = helper.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
		if err != nil {
			logrus.Errorf("cannot checkout branch: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	err = os.Remove(filepath.Join(tmpPath, envConfigFilePath))
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	hash, err := stageCommitAndPushWithHash(repo, tmpPath, token, fmt.Sprintf("[Gimlet] Deleting %s gimlet manifest for the %s env", configName, env))
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !pullRequestPolicy {
		response := map[string]interface{}{
			"link": fmt.Sprintf("%s/%s/commit/%s", dynamicConfig.ScmURL(), repoPath, hash),
		}

		responseJson, err := json.Marshal(response)
		if err != nil {
			logrus.Errorf("cannot marshal response: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseJson)
		return
	}

	createdPR, _, err := goScm.CreatePR(token, repoPath, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet] `%s` ➡️ `%s` deployment configuration change", configName, env),
		fmt.Sprintf("@%s is deleting the `%s` deployment configuration for the `%s` environment.", user.Login, configName, env))
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"link": createdPR.Link,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("cannot marshal manifest: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(repoPath)
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

// envConfigPath returns the envconfig file name based on convention, or the name of an existing file describing this env
func envConfigPath(env string, appName string, existingEnvConfigs map[string]*dx.Manifest) string {
	envConfigFileName := fmt.Sprintf("%s-%s.yaml", env, appName)
	for fileName, existingEnvConfig := range existingEnvConfigs {
		if existingEnvConfig.Env == env &&
			existingEnvConfig.App == appName {
			envConfigFileName = fileName
			break
		}
	}
	return envConfigFileName
}

func existingEnvConfigs(repo *git.Repository, headBranch string) (map[string]*dx.Manifest, error) {
	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, headBranch, ".gimlet")
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			return nil, fmt.Errorf("cannot list files in .gimlet/: %s", err)
		}
	}

	existingEnvConfigs := map[string]*dx.Manifest{}
	for fileName, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		existingEnvConfigs[fileName] = &envConfig
	}
	return existingEnvConfigs, nil
}

func getImportedRepos(dao *store.Store) ([]string, error) {
	var importedRepos []string
	importedReposJson, err := dao.KeyValue(model.ImportedRepos)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if importedReposJson.Value == "" {
		importedReposJson.Value = "[]"
	}

	err = json.Unmarshal([]byte(importedReposJson.Value), &importedRepos)
	if err != nil {
		return nil, err
	}

	return importedRepos, nil
}

func GenerateBranchNameWithUniqueHash(defaultBranchName string, uniqieHashlength int) (string, error) {
	b := make([]byte, uniqieHashlength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", defaultBranchName, hex.EncodeToString(b)), nil
}
