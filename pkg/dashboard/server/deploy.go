package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
)

func magicDeploy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	user := ctx.Value("user").(*model.User)
	// clientHub, _ := ctx.Value("clientHub").(*streaming.ClientHub)

	body, _ := ioutil.ReadAll(r.Body)
	var deployRequest dx.MagicDeployRequest
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&deployRequest)
	if err != nil {
		logrus.Errorf("cannot decode release request: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if deployRequest.Owner == "" || deployRequest.Repo == "" || deployRequest.Sha == "" {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "owner, repo, sha parameters are mandatory"), http.StatusBadRequest)
		return
	}

	envs, err := store.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if len(envs) != 1 {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed)+" - built-in environment missing", http.StatusPreconditionFailed)
		return
	}

	magicEnv := envs[0]

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	repo, repoPath, err := gitRepoCache.InstanceForWrite(deployRequest.Owner + "/" + deployRequest.Repo)
	defer os.RemoveAll(repoPath)
	if err != nil {
		logrus.Errorf("cannot get repo instance: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		logrus.Errorf("cannot get worktree: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(deployRequest.Sha),
		Mode:   git.HardReset,
	})
	if err != nil {
		logrus.Errorf("cannot set version: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		logrus.Errorf("cannot get temp file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tartar(tarFile.Name(), []string{repoPath})
	if err != nil {
		logrus.Errorf("cannot tar folder: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// config := ctx.Value("config").(*config.Config)

	imageBuildId := randStringRunes(6)
	// imageBuilderUrl := config.ImageBuilderHost + "/build-image"
	image := "registry.infrastructure.svc.cluster.local:5000/" + deployRequest.Repo
	tag := deployRequest.Sha

	imageBuilds, _ := ctx.Value("imageBuilds").(map[string]string)
	imageBuilds[imageBuildId] = tarFile.Name()

	agentHub, _ := ctx.Value("agentHub").(*streaming.AgentHub)
	agentHub.TriggerImageBuild(magicEnv.Name, imageBuildId, image, tag, deployRequest.Repo, user.Login)

	responseStr, err := json.Marshal(map[string]string{
		"buildId": string(imageBuildId),
	})
	if err != nil {
		logrus.Errorf("cannot serialize response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseStr)
}

// https://github.com/vladimirvivien/go-tar/blob/master/tartar/tartar.go
// tarrer walks paths to create tar file tarName
func tartar(tarName string, paths []string) (err error) {
	tarFile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer func() {
		err = tarFile.Close()
	}()

	absTar, err := filepath.Abs(tarName)
	if err != nil {
		return err
	}

	// enable compression if file ends in .gz
	tw := tar.NewWriter(tarFile)
	if strings.HasSuffix(tarName, ".gz") || strings.HasSuffix(tarName, ".gzip") {
		gz := gzip.NewWriter(tarFile)
		defer gz.Close()
		tw = tar.NewWriter(gz)
	}
	defer tw.Close()

	// walk each specified path and add encountered file to tar
	for _, path := range paths {
		// validate path
		path = filepath.Clean(path)
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if absPath == absTar {
			fmt.Printf("tar file %s cannot be the source\n", tarName)
			continue
		}
		if absPath == filepath.Dir(absTar) {
			fmt.Printf("tar file %s cannot be in source %s\n", tarName, absPath)
			continue
		}

		walker := func(file string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// fill in header info using func FileInfoHeader
			hdr, err := tar.FileInfoHeader(finfo, finfo.Name())
			if err != nil {
				return err
			}

			relFilePath := file
			if filepath.IsAbs(path) {
				relFilePath, err = filepath.Rel(path, file)
				if err != nil {
					return err
				}
			}
			// ensure header has relative file path
			hdr.Name = relFilePath

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			// if path is a dir, dont continue
			if finfo.Mode().IsDir() {
				return nil
			}

			// add file to tar
			srcFile, err := os.Open(file)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			_, err = io.Copy(tw, srcFile)
			if err != nil {
				return err
			}
			return nil
		}

		// build tar
		if err := filepath.Walk(path, walker); err != nil {
			fmt.Printf("failed to add %s to tar: %s\n", path, err)
		}
	}
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
