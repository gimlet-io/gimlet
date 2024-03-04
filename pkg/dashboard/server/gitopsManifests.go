package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func getGitopsManifests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	envName := chi.URLParam(r, "env")
	environment, err := store.GetEnvironment(envName)
	if err != nil {
		logrus.Errorf("cannot get env from db: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(environment.Name, "flux")
	if environment.RepoPerEnv {
		path = "flux"
	}

	infraRepoFiles, err := gitopsManifests(gitRepoCache, environment.InfraRepo, path)
	if err != nil {
		logrus.Errorf("cannot get gitops manifests from repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	appsRepoFiles, err := gitopsManifests(gitRepoCache, environment.AppsRepo, path)
	if err != nil {
		logrus.Errorf("cannot get gitops manifests from repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitopsManifests := map[string]interface{}{
		"infra": infraRepoFiles,
		"apps":  appsRepoFiles,
	}

	gitopsManifestsString, err := json.Marshal(gitopsManifests)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(gitopsManifestsString)
}

func gitopsManifests(gitRepoCache *nativeGit.RepoCache, repoName string, filesPath string) (map[string]string, error) {
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return nil, err
	}

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		return nil, err
	}

	files, err := nativeGit.RemoteFolderOnBranchWithoutCheckout(repo, headBranch, filesPath)
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			return nil, fmt.Errorf("cannot list files in %s: %s", filesPath, err)
		}
	}

	return files, nil
}
