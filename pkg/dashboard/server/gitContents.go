package server

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-chi/chi"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func branches(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	repoName := fmt.Sprintf("%s/%s", owner, name)

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(repoName)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	branches := []string{}
	refIter, _ := repo.References()
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
	repoName := chi.URLParam(r, "name")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	githubActionsConfigPath := filepath.Join(".github", "workflows")
	githubActionsShipperCommand := "gimlet-io/gimlet-artifact-shipper-action"
	hasGithubActionsConfig, hasGithubActionsShipper, err := hasCiConfigAndShipper(repo, githubActionsConfigPath, githubActionsShipperCommand)
	if err != nil {
		logrus.Errorf("cannot determine ci status: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	circleCiConfigPath := ".circleci"
	circleCiShipperCommand := "gimlet/gimlet-artifact-createn"
	hasCircleCiConfig, hasCircleCiShipper, err := hasCiConfigAndShipper(repo, circleCiConfigPath, circleCiShipperCommand)
	if err != nil {
		logrus.Errorf("cannot determine ci status: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	branch, err := helper.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, branch, ".gimlet")
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
		})
	}

	gitRepoM := gitRepoMetas{
		GithubActions: hasGithubActionsConfig,
		CircleCi:      hasCircleCiConfig,
		HasShipper:    hasGithubActionsShipper || hasCircleCiShipper,
		FileInfos:     fileInfos,
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

func getPullRequests(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	goScm := genericScm.NewGoScmHelper(config, nil)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	db := r.Context().Value("store").(*store.Store)
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get all environments from database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	prList, err := goScm.ListOpenPRs(token, repoPath)
	if err != nil {
		logrus.Errorf("cannot list pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var prListCreatedByGimlet []*scm.PullRequest
	for _, pullRequest := range prList {
		if strings.HasPrefix(pullRequest.Source, "gimlet-config-change") {
			prListCreatedByGimlet = append(prListCreatedByGimlet, pullRequest)
		}
	}

	pullRequests := map[string]interface{}{}
	for _, env := range envsFromDB {
		var pullRequestsByEnv []*api.PR
		for _, pullRequest := range prListCreatedByGimlet {
			if strings.Contains(pullRequest.Source, env.Name) {
				pullRequestsByEnv = append(pullRequestsByEnv, &api.PR{
					Sha:     pullRequest.Sha,
					Link:    pullRequest.Link,
					Title:   pullRequest.Title,
					Source:  pullRequest.Source,
					Number:  pullRequest.Number,
					Author:  pullRequest.Author.Login,
					Created: int(pullRequest.Created.Unix()),
					Updated: int(pullRequest.Updated.Unix()),
				})
			}
		}

		pullRequests[env.Name] = pullRequestsByEnv
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

func getPullRequestsFromInfraRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	goScm := genericScm.NewGoScmHelper(config, nil)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	db := r.Context().Value("store").(*store.Store)
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get all environments from database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	infraRepoPullRequests := map[string]interface{}{}
	for _, env := range envsFromDB {
		prList, err := goScm.ListOpenPRs(token, env.InfraRepo)
		if err != nil {
			if !strings.Contains(err.Error(), "Not Found") {
				logrus.Errorf("cannot list pull requests: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		var prListCreatedByGimlet []*api.PR
		for _, pullRequest := range prList {
			if strings.HasPrefix(pullRequest.Source, "gimlet-stack-change") && strings.Contains(pullRequest.Source, env.Name) {
				prListCreatedByGimlet = append(prListCreatedByGimlet, &api.PR{
					Sha:     pullRequest.Sha,
					Link:    pullRequest.Link,
					Title:   pullRequest.Title,
					Source:  pullRequest.Source,
					Number:  pullRequest.Number,
					Author:  pullRequest.Author.Login,
					Created: int(pullRequest.Created.Unix()),
					Updated: int(pullRequest.Updated.Unix()),
				})
			}
		}

		infraRepoPullRequests[env.Name] = prListCreatedByGimlet
	}

	infraRepoPullRequestsString, err := json.Marshal(infraRepoPullRequests)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(infraRepoPullRequestsString)
}

type fileInfo struct {
	EnvName  string `json:"envName"`
	AppName  string `json:"appName"`
	FileName string `json:"fileName"`
}

type gitRepoMetas struct {
	GithubActions bool       `json:"githubActions"`
	CircleCi      bool       `json:"circleCi"`
	HasShipper    bool       `json:"hasShipper"`
	FileInfos     []fileInfo `json:"fileInfos"`
}

// envConfig fetches all environment configs from source control for a repo
func envConfigs(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	branch, err := helper.HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, branch, ".gimlet")
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

type envConfig struct {
	Values          map[string]interface{}
	Namespace       string
	Chart           Chart
	AppName         string
	UseDeployPolicy bool
	DeployBranch    string
	DeployTag       string
	DeployEvent     *dx.GitEvent
}

type Chart struct {
	Repository string
	Name       string
	Version    string
}

func saveEnvConfig(w http.ResponseWriter, r *http.Request) {
	envConfigData := &envConfig{}
	err := json.NewDecoder(r.Body).Decode(&envConfigData)
	if err != nil {
		logrus.Errorf("cannot decode env config data: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)
	env := chi.URLParam(r, "env")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	config := ctx.Value("config").(*config.Config)
	user := ctx.Value("user").(*model.User)
	goScm := genericScm.NewGoScmHelper(config, nil)

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

	envConfigFileName := envConfigPath(env, envConfigData, existingEnvConfigs)
	envConfigFilePath := fmt.Sprintf(".gimlet/%s", envConfigFileName)

	// Temporary solution for some encoding problem (https://github.com/remix-run/history/issues/505), which fixed in https://github.com/remix-run/history/releases/tag/v5.0.0.
	// Must update react-router-dom to v6 in the near future.
	envConfigFilePath = strings.ReplaceAll(envConfigFilePath, "%", "%25")

	createCase := false // indicates if we need to create the file, we update existing manifests on the default path
	_, _, err = goScm.Content(token, repoPath, envConfigFilePath, headBranch)
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

	var manifest *dx.Manifest
	if createCase {
		manifest = &dx.Manifest{
			App: envConfigData.AppName,
			Env: env,
			Chart: dx.Chart{
				Name:       envConfigData.Chart.Name,
				Repository: envConfigData.Chart.Repository,
				Version:    envConfigData.Chart.Version,
			},
			Namespace: envConfigData.Namespace,
			Values:    envConfigData.Values,
		}
	} else { // we are updating an existing manifest file
		manifest = existingEnvConfigs[envConfigFileName]
		manifest.Values = envConfigData.Values
		manifest.Namespace = envConfigData.Namespace
	}

	if !createCase && manifest.App != envConfigData.AppName { // we do not allow updating application names
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if envConfigData.UseDeployPolicy {
		manifest.Deploy = &dx.Deploy{
			Branch: envConfigData.DeployBranch,
			Event:  envConfigData.DeployEvent,
			Tag:    envConfigData.DeployTag,
		}
	} else {
		manifest.Deploy = nil
	}

	// marshall and ident the manifest
	var toSaveBuffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&toSaveBuffer)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&manifest)
	if err != nil {
		logrus.Errorf("cannot marshal manifest: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// generate branch name to write changes on
	sourceBranch, err := generateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-config-change-%s", env), 4)
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

	message := fmt.Sprintf("[Gimlet Dashboard] Updating %s gimlet manifest for the %s env", envConfigData.AppName, env)
	if createCase {
		message = fmt.Sprintf("[Gimlet Dashboard] Creating %s gimlet manifest for the %s env", envConfigData.AppName, env)
	}

	_ = os.MkdirAll(filepath.Join(tmpPath, ".gimlet"), nativeGit.Dir_RWX_RX_R)
	err = os.WriteFile(filepath.Join(tmpPath, envConfigFilePath), toSaveBuffer.Bytes(), nativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = StageCommitAndPush(repo, tmpPath, token, message)
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	createdPR, _, err := goScm.CreatePR(token, repoPath, sourceBranch, headBranch,
		fmt.Sprintf("[Gimlet Dashboard] `%s` ➡️ `%s` deployment configuration change", envConfigData.AppName, env),
		fmt.Sprintf("@%s is editing the `%s` deployment configuration for the `%s` environment.", user.Login, envConfigData.AppName, env))
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"envName": env,
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
		"manifest": manifest,
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
func envConfigPath(env string, envConfigData *envConfig, existingEnvConfigs map[string]*dx.Manifest) string {
	envConfigFileName := fmt.Sprintf("%s-%s.yaml", env, envConfigData.AppName)
	for fileName, existingEnvConfig := range existingEnvConfigs {
		if existingEnvConfig.Env == env &&
			existingEnvConfig.App == envConfigData.AppName {
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

func getOrgRepos(dao *store.Store) ([]string, error) {
	var orgRepos []string
	orgReposJson, err := dao.KeyValue(model.OrgRepos)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if orgReposJson.Value == "" {
		orgReposJson.Value = "[]"
	}

	err = json.Unmarshal([]byte(orgReposJson.Value), &orgRepos)
	if err != nil {
		return nil, err
	}

	return orgRepos, nil
}

func hasRepo(orgRepos []string, repo string) bool {
	for _, orgRepo := range orgRepos {
		if orgRepo == repo {
			return true
		}
	}
	return false
}

func hasShipper(files map[string]string, shipperCommand string) bool {
	for _, file := range files {
		if strings.Contains(file, shipperCommand) {
			return true
		}
	}
	return false
}

func hasCiConfigAndShipper(repo *git.Repository, ciConfigPath string, shipperCommand string) (bool, bool, error) {
	branch, err := helper.HeadBranch(repo)
	if err != nil {
		return false, false, err
	}

	ciConfigFiles, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, branch, ciConfigPath)
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			return false, false, err
		}
	}

	if len(ciConfigFiles) == 0 {
		return false, false, nil
	}

	return true, hasShipper(ciConfigFiles, shipperCommand), nil
}

func generateBranchNameWithUniqueHash(defaultBranchName string, uniqieHashlength int) (string, error) {
	b := make([]byte, uniqieHashlength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", defaultBranchName, hex.EncodeToString(b)), nil
}
