package server

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
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

	branch := helper.HeadBranch(repo)
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
	pullRequests["repo"] = repoPath
	pullRequests["prList"] = prListCreatedByGimlet

	pullRequestsString, err := json.Marshal(pullRequests)
	if err != nil {
		logrus.Errorf("cannot serialize pull requests: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(pullRequestsString)
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

	branch := helper.HeadBranch(repo)

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
	AppName         string
	UseDeployPolicy bool
	DeployBranch    string
	DeployTag       string
	DeployEvent     int
}

func saveEnvConfig(w http.ResponseWriter, r *http.Request) {
	envConfigData := &envConfig{}
	err := json.NewDecoder(r.Body).Decode(&envConfigData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
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
	goScm := genericScm.NewGoScmHelper(config, nil)

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	headBranch := helper.HeadBranch(repo)
	existingEnvConfigs, err := existingEnvConfigs(repo, headBranch)
	if err != nil {
		logrus.Errorf(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	envConfigFileName := envConfigPath(env, envConfigData, existingEnvConfigs)
	envConfigFilePath := fmt.Sprintf(".gimlet/%s", envConfigFileName)

	createCase := false // indicates if we need to create the file, we update existing manifests on the default path
	_, blobID, err := goScm.Content(token, repoPath, envConfigFilePath, headBranch)
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
				Name:       config.Chart.Name,
				Repository: config.Chart.Repo,
				Version:    config.Chart.Version,
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

	event := dx.GitEvent(envConfigData.DeployEvent)
	if envConfigData.UseDeployPolicy {
		manifest.Deploy = &dx.Deploy{
			Branch: envConfigData.DeployBranch,
			Event:  &event,
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

	// create branch to write changes on
	sourceBranch, err := generateBranchNameWithUniqueHash(fmt.Sprintf("gimlet-config-change-%s", env), 4)
	if err != nil {
		logrus.Errorf("cannot create branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ref, err := repo.Head()
	if err != nil {
		logrus.Errorf("cannot get head: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = goScm.CreateBranch(token, repoPath, sourceBranch, ref.Hash().String())
	if err != nil {
		logrus.Errorf("cannot create branch: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if createCase {
		err = goScm.CreateContent(
			token,
			repoPath,
			envConfigFilePath,
			toSaveBuffer.Bytes(),
			sourceBranch,
			fmt.Sprintf("[Gimlet Dashboard] Creating %s gimlet manifest for the %s env", envConfigData.AppName, env),
		)
		if err != nil {
			logrus.Errorf("cannot write git: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		err = goScm.UpdateContent(
			token,
			repoPath,
			envConfigFilePath,
			toSaveBuffer.Bytes(),
			blobID,
			sourceBranch,
			fmt.Sprintf("[Gimlet Dashboard] Updating %s gimlet manifest for the %s env", envConfigData.AppName, env),
		)
		if err != nil {
			logrus.Errorf("cannot write git: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	_, _, err = goScm.CreatePR(token, repoPath, sourceBranch, headBranch, "[Gimlet Dashboard] Environment config change", "Gimlet Dashboard has created this PR")
	if err != nil {
		logrus.Errorf("cannot create pr: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	manifestJson, err := json.Marshal(manifest)
	if err != nil {
		logrus.Errorf("cannot marshal manifest: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(repoPath)
	w.WriteHeader(http.StatusOK)
	w.Write(manifestJson)
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
	branch := helper.HeadBranch(repo)
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
