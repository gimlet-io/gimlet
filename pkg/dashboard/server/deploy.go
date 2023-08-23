package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func magicDeploy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	user := ctx.Value("user").(*model.User)

	body, _ := ioutil.ReadAll(r.Body)
	var deployRequest dx.MagicDeployRequest
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&deployRequest)
	if err != nil {
		logrus.Errorf("cannot decode release request: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	deployRequest.TriggeredBy = user.Login

	if deployRequest.Owner == "" || deployRequest.Repo == "" || deployRequest.Sha == "" {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "owner, repo, sha parameters are mandatory"), http.StatusBadRequest)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	repo, repoPath, err := gitRepoCache.InstanceForWrite(deployRequest.Owner + "/" + deployRequest.Repo)
	defer os.RemoveAll(repoPath)
	if err != nil {
		logrus.Errorf("cannot get repo instance: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	envConfig, err := gitops.Manifest(repo, deployRequest.Sha, deployRequest.Env, deployRequest.App)
	if err != nil {
		logrus.Errorf("cannot get repo instance: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var imageBuildId string
	if envConfig != nil && envConfig.Chart.Name == "static-site" {
		imageBuildId = "static-" + randStringRunes(6)

		version, err := gitops.Version(deployRequest.Owner, deployRequest.Repo, repo, deployRequest.Sha)
		if err != nil {
			logrus.Errorf("cannot get version: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		clientHub, _ := ctx.Value("clientHub").(*streaming.ClientHub)

		go createDummyArtifactAndStreamToClient(
			deployRequest,
			version,
			envConfig,
			imageBuildId,
			store,
			clientHub,
		)
	} else {
		imageBuildId, err = triggerImageBuild(repo, repoPath, deployRequest, ctx)
		if err != nil {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

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

func triggerImageBuild(
	repo *git.Repository,
	repoPath string,
	deployRequest dx.MagicDeployRequest,
	ctx context.Context,
) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("cannot get worktree: %s", err)
	}
	err = worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(deployRequest.Sha),
		Mode:   git.HardReset,
	})
	if err != nil {
		return "", fmt.Errorf("cannot set version: %s", err)
	}

	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("cannot get temp file: %s", err)
	}

	err = tartar(tarFile.Name(), []string{repoPath})
	if err != nil {
		return "", fmt.Errorf("cannot tar folder: %s", err)
	}

	imageBuildId := randStringRunes(6)
	image := "registry.infrastructure.svc.cluster.local:5000/" + deployRequest.App
	tag := deployRequest.Sha

	trigger := streaming.ImageBuildTrigger{
		DeployRequest: deployRequest,
		ImageBuildId:  imageBuildId,
		Image:         image,
		Tag:           tag,
		SourcePath:    tarFile.Name(),
	}

	imageBuilds, _ := ctx.Value("imageBuilds").(map[string]streaming.ImageBuildTrigger)
	imageBuilds[imageBuildId] = trigger

	agentHub, _ := ctx.Value("agentHub").(*streaming.AgentHub)
	agentHub.TriggerImageBuild(trigger)

	return imageBuildId, nil
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

func createDummyArtifactAndStreamToClient(
	deployRequest dx.MagicDeployRequest,
	version *dx.Version,
	manifest *dx.Manifest,
	imageBuildId string,
	store *store.Store,
	clientHub *streaming.ClientHub,
) {
	time.Sleep(time.Millisecond * 200) // wait til client learns about the buildID
	artifact := &dx.Artifact{
		ID:           fmt.Sprintf("%s-%s", deployRequest.Owner+"/"+deployRequest.Repo, uuid.New().String()),
		Created:      time.Now().Unix(),
		Fake:         true,
		Environments: []*dx.Manifest{manifest},
		Version:      *version,
		Vars: map[string]string{
			"SHA": deployRequest.Sha,
		},
	}

	event, err := model.ToEvent(*artifact)
	if err != nil {
		logrus.Errorf("cannot convert to artifact model: %s", err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}

	_, err = store.CreateEvent(event)
	if err != nil {
		logrus.Errorf("cannot save artifact: %s", err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         deployRequest.Env,
		App:         deployRequest.App,
		ArtifactID:  artifact.ID,
		TriggeredBy: deployRequest.TriggeredBy,
	})
	if err != nil {
		logrus.Errorf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}

	artifactEvent, err := store.Artifact(artifact.ID)
	if err != nil {
		logrus.Errorf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), artifact.ID)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}
	event, err = store.CreateEvent(&model.Event{
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: artifactEvent.Repository,
	})
	if err != nil {
		logrus.Errorf("%s - cannot save release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}

	streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "created", event.ID)

}

func streamArtifactCreatedEvent(clientHub *streaming.ClientHub, userLogin string, imageBuildId string, status string, trackingId string) {
	jsonString, _ := json.Marshal(streaming.ArtifactCreatedEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.ArtifactCreatedEventString},
		BuildId:        imageBuildId,
		TrackingId:     trackingId,
	})
	clientHub.Send <- &streaming.ClientMessage{
		ClientId: userLogin,
		Message:  jsonString,
	}
}
