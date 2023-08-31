package worker

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/sirupsen/logrus"
)

type ImageBuildWorker struct {
	gitRepoCache           *nativeGit.RepoCache
	clientHub              *streaming.ClientHub
	store                  *store.Store
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage
	imageBuilds            map[string]streaming.ImageBuildTrigger
}

func NewImageBuildWorker(
	gitRepoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	store *store.Store,
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage,
	imageBuilds map[string]streaming.ImageBuildTrigger,
) *ImageBuildWorker {
	imageBuildWorker := &ImageBuildWorker{
		gitRepoCache:           gitRepoCache,
		clientHub:              clientHub,
		store:                  store,
		successfullImageBuilds: successfullImageBuilds,
		imageBuilds:            imageBuilds,
	}

	return imageBuildWorker
}

func (m *ImageBuildWorker) Run() {
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
				)
			}

			if imageBuildStatus.Status != "running" {
				os.RemoveAll(imageBuild.SourcePath)
				delete(m.imageBuilds, imageBuildStatus.BuildId)
			}
		}
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

func createDeployRequest(
	deployRequest dx.MagicDeployRequest,
	store *store.Store,
	tag string,
	clientHub *streaming.ClientHub,
	imageBuildId string,
	gitRepoCache *nativeGit.RepoCache,
) {
	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         deployRequest.Env,
		App:         deployRequest.App,
		ArtifactID:  deployRequest.ArtifactID,
		TriggeredBy: deployRequest.TriggeredBy,
	})
	if err != nil {
		logrus.Errorf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}

	artifactEvent, err := store.Artifact(deployRequest.ArtifactID)
	if err != nil {
		logrus.Errorf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), deployRequest.ArtifactID)
		streamArtifactCreatedEvent(clientHub, deployRequest.TriggeredBy, imageBuildId, "error", "")
		return
	}
	event, err := store.CreateEvent(&model.Event{
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
