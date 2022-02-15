package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	helper "github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/nativeGit"
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
		if r.Name().IsBranch() {
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

	files, err := helper.Folder(repo, ".gimlet")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
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

func saveEnvConfig(w http.ResponseWriter, r *http.Request) {
	var values map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&values)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)

	env := chi.URLParam(r, "env")
	configName := chi.URLParam(r, "config")

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	files, err := helper.Folder(repo, ".gimlet")
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
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

	fileToUpdate := fmt.Sprintf("%s.yaml", env)
	for fileName, existingEnvConfig := range existingEnvConfigs {
		if existingEnvConfig.Env == env &&
			existingEnvConfig.App == configName {
			fileToUpdate = fileName
			break
		}
	}
	fileUpdatePath := fmt.Sprintf(".gimlet/%s", fileToUpdate)

	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	config := ctx.Value("config").(*config.Config)
	goScm := genericScm.NewGoScmHelper(config, nil)

	branch := headBranch(repo)

	_, blobID, err := goScm.Content(token, repoPath, fileUpdatePath, branch)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {

			toSave := &dx.Manifest{
				App: configName,
				Env: env,
				Chart: dx.Chart{
					Name:       "onechart",
					Repository: "https://chart.onechart.dev",
					Version:    "0.32.0",
				},
				Namespace: "staging",
				Values:    values,
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
				fmt.Sprintf("[Gimlet Dashboard] Creating %s gimlet manifest for the %s env", env, configName),
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
		toUpdate.Values = values

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
			fmt.Sprintf("[Gimlet Dashboard] Updating %s gimlet manifest for the %s env", env, configName),
		)
		if err != nil {
			logrus.Errorf("cannot update manifest: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gitRepoCache.Invalidate(repoPath)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func branchList(repo *git.Repository) []string {
	branches := []string{}
	refIter, _ := repo.References()
	refIter.ForEach(func(r *plumbing.Reference) error {
		if r.Name().IsBranch() {
			branch := r.Name().Short()
			branches = append(branches, strings.TrimPrefix(branch, "origin/"))
		}
		return nil
	})

	return branches
}

func headBranch(repo *git.Repository) string {
	branches := branchList(repo)
	for _, b := range branches {
		if b == "main" {
			return "main"
		}
	}
	return "master"
}
