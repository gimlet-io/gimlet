package server

import (
	"bytes"
	"database/sql"
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
	Values    map[string]interface{}
	Namespace string
	AppName   string
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

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	headBranch := helper.HeadBranch(repo)
	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, headBranch, ".gimlet")
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			logrus.Errorf("cannot list files in .gimlet/: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	existingEnvConfigs := map[string]dx.Manifest{}
	for fileName, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		existingEnvConfigs[fileName] = envConfig
	}

	fileToUpdate := fmt.Sprintf("%s-%s.yaml", env, envConfigData.AppName)
	for fileName, existingEnvConfig := range existingEnvConfigs {
		if existingEnvConfig.Env == env &&
			existingEnvConfig.App == envConfigData.AppName {
			fileToUpdate = fileName
			break
		}
	}
	fileUpdatePath := fmt.Sprintf(".gimlet/%s", fileToUpdate)

	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	config := ctx.Value("config").(*config.Config)
	goScm := genericScm.NewGoScmHelper(config, nil)

	branch := helper.HeadBranch(repo)

	_, blobID, err := goScm.Content(token, repoPath, fileUpdatePath, branch)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {

			toSave := &dx.Manifest{
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

			var toSaveBuffer bytes.Buffer
			yamlEncoder := yaml.NewEncoder(&toSaveBuffer)
			yamlEncoder.SetIndent(2)
			err = yamlEncoder.Encode(&toSave)

			if err != nil {
				logrus.Errorf("cannot marshal manifest: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			err = goScm.CreateContent(
				token,
				repoPath,
				fileUpdatePath,
				toSaveBuffer.Bytes(),
				branch,
				fmt.Sprintf("[Gimlet Dashboard] Creating %s gimlet manifest for the %s env", envConfigData.AppName, env),
			)
			if err != nil {
				logrus.Errorf("cannot create manifest: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			logrus.Errorf("cannot fetch envConfig from github: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			w.Write([]byte("{}"))
			return
		}
	} else {
		toUpdate := existingEnvConfigs[fileToUpdate]
		toUpdate.Values = envConfigData.Values
		toUpdate.Namespace = envConfigData.Namespace

		if toUpdate.App != envConfigData.AppName {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var toUpdateBuffer bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&toUpdateBuffer)
		yamlEncoder.SetIndent(2)
		err = yamlEncoder.Encode(&toUpdate)
		if err != nil {
			logrus.Errorf("cannot marshal manifest: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = goScm.UpdateContent(
			token,
			repoPath,
			fileUpdatePath,
			toUpdateBuffer.Bytes(),
			blobID,
			branch,
			fmt.Sprintf("[Gimlet Dashboard] Updating %s gimlet manifest for the %s env", env, envConfigData.AppName),
		)
		if err != nil {
			logrus.Errorf("cannot update manifest: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	gitRepoCache.Invalidate(repoPath)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
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

func folderExists(gitopsRepoContent []string, envName string) bool {
	for _, content := range gitopsRepoContent {
		if content == envName {
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
