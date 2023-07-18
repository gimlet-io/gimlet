package server

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func magicDeploy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	// user := ctx.Value("user").(*model.User)
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
	agentHub.TriggerImageBuild(magicEnv.Name, imageBuildId, image, tag)

	// signalCh := make(chan imageBuildingDoneSignal)
	// go buildImage(
	// 	tarFile.Name(),
	// 	imageBuilderUrl,
	// 	image,
	// 	tag,
	// 	deployRequest.Repo,
	// 	signalCh,
	// 	clientHub,
	// 	user.Login,
	// 	string(imageBuildId),
	// )

	// go createDeployRequest(
	// 	deployRequest,
	// 	magicEnv,
	// 	store,
	// 	tag,
	// 	user.Login,
	// 	signalCh,
	// 	clientHub,
	// 	user.Login,
	// 	string(imageBuildId),
	// 	gitRepoCache,
	// 	magicEnv.Name,
	// )

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

type imageBuildingDoneSignal struct {
	successful bool
}

func createDeployRequest(
	deployRequest dx.MagicDeployRequest,
	builtInEnv *model.Environment,
	store *store.Store,
	tag string,
	triggeredBy string,
	signal chan imageBuildingDoneSignal,
	clientHub *streaming.ClientHub,
	userLogin string,
	imageBuildId string,
	gitRepoCache *nativeGit.RepoCache,
	builtInEnvName string,
) {
	// wait until image building is done
	imageBuildingDoneSignal := <-signal
	if !imageBuildingDoneSignal.successful {
		return
	}
	envConfig, _ := defaultEnvConfig(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha, builtInEnvName,
		gitRepoCache,
	)

	artifact, err := createDummyArtifact(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha,
		builtInEnv.Name,
		store,
		"127.0.0.1:32447/"+deployRequest.Repo,
		tag,
		envConfig,
	)
	if err != nil {
		logrus.Errorf("cannot create artifact: %s", err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         builtInEnv.Name,
		App:         deployRequest.Repo,
		ArtifactID:  artifact.ID,
		TriggeredBy: triggeredBy,
	})
	if err != nil {
		logrus.Errorf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	artifactEvent, err := store.Artifact(artifact.ID)
	if err != nil {
		logrus.Errorf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), artifact.ID)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}
	event, err := store.CreateEvent(&model.Event{
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: artifactEvent.Repository,
	})
	if err != nil {
		logrus.Errorf("%s - cannot save release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "created", event.ID)
}

func defaultEnvConfig(
	owner string, repoName string, sha string, env string,
	gitRepoCache *nativeGit.RepoCache,

) (*dx.Manifest, error) {
	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		return nil, fmt.Errorf("cannot get repo: %s", err)
	}

	files, err := helper.RemoteFolderOnHashWithoutCheckout(repo, sha, ".gimlet")
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			return nil, nil
		} else {
			return nil, fmt.Errorf("cannot list files in .gimlet/: %s", err)
		}
	}

	for _, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		if envConfig.Env == env && envConfig.App == repoName {
			return &envConfig, nil
		}
	}

	return nil, nil
}

func createDummyArtifact(
	owner, repo, sha string,
	env string,
	store *store.Store,
	image, tag string,
	envConfig *dx.Manifest,
) (*dx.Artifact, error) {

	if envConfig == nil {
		envConfig = &dx.Manifest{
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
					"pullPolicy": "Always",
				},
				"resources": map[string]interface{}{
					"ignore": true,
				},
			},
		}
	}

	artifact := dx.Artifact{
		ID:           fmt.Sprintf("%s-%s", owner+"/"+repo, uuid.New().String()),
		Created:      time.Now().Unix(),
		Fake:         true,
		Environments: []*dx.Manifest{envConfig},
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
		Vars: map[string]string{
			"SHA": sha,
		},
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

func buildImage(
	path string, url string,
	image string, tag string, app string,
	signalCh chan imageBuildingDoneSignal,
	clientHub *streaming.ClientHub,
	userLogin string,
	imageBuildId string,
) {
	request, err := newfileUploadRequest(url, map[string]string{
		"image": image,
		"tag":   tag,
		"app":   app,
	}, "data", path)
	if err != nil {
		signalCh <- imageBuildingDoneSignal{successful: false}
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		signalCh <- imageBuildingDoneSignal{successful: false}
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	success := streamImageBuilderLogs(resp.Body, clientHub, userLogin, imageBuildId)
	signalCh <- imageBuildingDoneSignal{successful: success}
}

func streamImageBuilderLogs(body io.ReadCloser, clientHub *streaming.ClientHub, userLogin string, imageBuildId string) bool {
	defer body.Close()
	var sb strings.Builder
	reader := bufio.NewReader(body)
	first := true
	for {
		line, err := reader.ReadBytes('\n')
		sb.WriteString(string(line))
		if err != nil {
			if err == io.EOF {
				break
			}

			logrus.Errorf("cannot stream build logs: %s", err)
			streamImageBuildEvent(clientHub, userLogin, imageBuildId, "error", sb.String())
			return false
		}

		if first || sb.Len() > 1000 {
			streamImageBuildEvent(clientHub, userLogin, imageBuildId, "running", sb.String())
			sb.Reset()
			first = false
		}
	}

	lastLine := sb.String()
	if strings.HasSuffix(lastLine, "IMAGE BUILT") {
		streamImageBuildEvent(clientHub, userLogin, imageBuildId, "success", lastLine)
		return true
	} else {
		streamImageBuildEvent(clientHub, userLogin, imageBuildId, "notBuilt", lastLine)
		return false
	}
}

func streamImageBuildEvent(clientHub *streaming.ClientHub, userLogin string, imageBuildId string, status string, logLine string) {
	jsonString, _ := json.Marshal(streaming.ImageBuildLogEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.ImageBuildLogEventString},
		BuildId:        imageBuildId,
		Status:         status,
		LogLine:        string(logLine),
	})
	clientHub.Send <- &streaming.ClientMessage{
		ClientId: userLogin,
		Message:  jsonString,
	}
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
