package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
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

	var builtInEnv *model.Environment
	for _, env := range envs {
		if env.BuiltIn {
			builtInEnv = env
			break
		}
	}
	if builtInEnv == nil {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed)+" - built-in environment missing", http.StatusPreconditionFailed)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	_, repoPath, err := gitRepoCache.InstanceForWrite(deployRequest.Owner + "/" + deployRequest.Repo)
	defer os.RemoveAll(repoPath)
	if err != nil {
		logrus.Errorf("cannot get repo instance: %s", err)
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

	imageBuilderUrl := "http://127.0.0.1:8000/build-image"
	image := "registry.acorn-image-system.svc.cluster.local:5000/" + deployRequest.Repo
	// previousImage := "" //"registry.acorn-image-system.svc.cluster.local:5000/dummy"
	tag := deployRequest.Sha
	err = buildImage(tarFile.Name(), imageBuilderUrl, image, tag)
	if err != nil {
		logrus.Errorf("cannot tar folder: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	artifact, err := createDummyArtifact(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha,
		builtInEnv.Name,
		store,
		"127.0.0.1:31845/"+deployRequest.Repo,
		tag,
	)
	if err != nil {
		logrus.Errorf("cannot create artifact: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         builtInEnv.Name,
		App:         deployRequest.Repo,
		ArtifactID:  artifact.ID,
		TriggeredBy: user.Login,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
		return
	}

	artifactEvent, err := store.Artifact(artifact.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), artifact.ID), http.StatusNotFound)
		return
	}
	event, err := store.CreateEvent(&model.Event{
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: artifactEvent.Repository,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot save release request: %s", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
		return
	}

	eventIDBytes, _ := json.Marshal(map[string]string{
		"id": event.ID,
	})

	w.WriteHeader(http.StatusCreated)
	w.Write(eventIDBytes)
}

func createDummyArtifact(
	owner, repo, sha string,
	env string,
	store *store.Store,
	image, tag string,
) (*dx.Artifact, error) {
	artifact := dx.Artifact{
		ID:      fmt.Sprintf("%s-%s", owner+"/"+repo, uuid.New().String()),
		Created: time.Now().Unix(),
		Fake:    true,
		Environments: []*dx.Manifest{
			{
				App:       repo,
				Namespace: "default",
				Env:       env,
				Chart: dx.Chart{
					Name:       config.DEFAULT_CHART_NAME,
					Repository: config.DEFAULT_CHART_REPO,
					Version:    config.DEFAULT_CHART_VERSION,
				},
				Values: map[string]interface{}{
					"containerPort": 80,
					"gitRepository": owner + "/" + repo,
					"gitSha":        sha,
					"image": map[string]interface{}{
						"repository": image,
						"tag":        tag,
					},
				},
			},
		},
		Version: dx.Version{
			RepositoryName: owner + "/" + repo,
			SHA:            sha,
			Created:        time.Now().Unix(),
			Branch:         "main",
			AuthorName:     "TODO",
			AuthorEmail:    "TODO",
			CommitterName:  "TODO",
			CommitterEmail: "TODO",
			Message:        "TODO",
			URL:            "TODO",
		},
		Vars: map[string]string{},
	}

	event, err := model.ToEvent(artifact)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to artifact model: %s", err)
	}

	_, err = store.CreateEvent(event)
	if err != nil {
		return nil, fmt.Errorf("cannot save artifact: %s", err)
	}

	return &artifact, nil
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

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func buildImage(path string, url string, image string, tag string) error {
	request, err := newfileUploadRequest(url, map[string]string{
		"image": image,
		"tag":   tag,
	}, "data", path)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("image builder returned %d: %s", resp.StatusCode, body)
		}
	}

	return nil
}
