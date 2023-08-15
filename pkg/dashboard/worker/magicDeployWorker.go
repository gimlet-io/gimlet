package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type MagicDeployWorker struct {
	gitRepoCache           *nativeGit.RepoCache
	clientHub              *streaming.ClientHub
	store                  *store.Store
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage
	imageBuilds            map[string]streaming.ImageBuildTrigger
}

func NewMagicDeployWorker(
	gitRepoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	store *store.Store,
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage,
	imageBuilds map[string]streaming.ImageBuildTrigger,
) *MagicDeployWorker {
	magicDeployWorker := &MagicDeployWorker{
		gitRepoCache:           gitRepoCache,
		clientHub:              clientHub,
		store:                  store,
		successfullImageBuilds: successfullImageBuilds,
		imageBuilds:            imageBuilds,
	}

	return magicDeployWorker
}

func (m *MagicDeployWorker) Run() {
	for {
		select {
		case imageBuildStatus := <-m.successfullImageBuilds:
			imageBuild := m.imageBuilds[imageBuildStatus.BuildId]

			if imageBuildStatus.Status == "success" {
				go createDeployRequest(
					imageBuild.DeployRequest,
					m.store,
					imageBuild.Tag,
					m.clientHub,
					string(imageBuild.ImageBuildId),
					m.gitRepoCache,
					imageBuild.Env,
				)
			}

			if imageBuildStatus.Status != "running" {
				os.RemoveAll(imageBuild.SourcePath)
				delete(m.imageBuilds, imageBuildStatus.BuildId)
			}
		}
	}
}
func createDummyArtifact(
	owner, repo, sha string,
	env string,
	image, tag string,
	envConfig *dx.Manifest,
	version dx.Version,
) (*dx.Artifact, error) {
	defaultChart, err := config.DefaultChart()
	if err != nil {
		return nil, fmt.Errorf("cannot get default chart from config: %s", err)
	}

	if envConfig == nil {
		envConfig = &dx.Manifest{
			App:       repo,
			Namespace: "default",
			Env:       env,
			Chart:     *defaultChart,
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
		Version:      version,
		Vars: map[string]string{
			"SHA": sha,
		},
	}

	return &artifact, nil
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

func createDeployRequest(
	deployRequest dx.MagicDeployRequest,
	store *store.Store,
	tag string,
	clientHub *streaming.ClientHub,
	imageBuildId string,
	gitRepoCache *nativeGit.RepoCache,
	builtInEnvName string,
) {
	envConfig, version, _ := defaultEnvConfig(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha, builtInEnvName,
		gitRepoCache,
	)

	artifact, err := createDummyArtifact(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha,
		builtInEnvName,
		"127.0.0.1:32447/"+deployRequest.Repo,
		tag,
		envConfig,
		version,
	)
	if err != nil {
		logrus.Errorf("cannot create artifact: %s", err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
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
		Env:         builtInEnvName,
		App:         deployRequest.Repo,
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

func defaultEnvConfig(
	owner string, repoName string, sha string, env string,
	gitRepoCache *nativeGit.RepoCache,

) (*dx.Manifest, dx.Version, error) {
	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		return nil, dx.Version{}, fmt.Errorf("cannot get repo: %s", err)
	}

	c, err := repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return nil, dx.Version{}, fmt.Errorf("cannot get commit: %s", err)
	}
	version := dx.Version{
		RepositoryName: owner + "/" + repoName,
		SHA:            sha,
		Created:        time.Now().Unix(),
		Branch:         "TODO",
		AuthorName:     c.Author.Name,
		AuthorEmail:    c.Author.Email,
		CommitterName:  c.Committer.Name,
		CommitterEmail: c.Committer.Email,
		Message:        c.Message,
		URL:            "TODO",
	}

	files, err := nativeGit.RemoteFolderOnHashWithoutCheckout(repo, sha, ".gimlet")
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			return nil, version, nil
		} else {
			return nil, version, fmt.Errorf("cannot list files in .gimlet/: %s", err)
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
			return &envConfig, version, nil
		}
	}

	return nil, version, nil
}
