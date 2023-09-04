package worker

import (
	"encoding/json"
	"os"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/sirupsen/logrus"
)

type ImageBuildWorker struct {
	store                  *store.Store
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage
}

func NewImageBuildWorker(
	store *store.Store,
	successfullImageBuilds chan streaming.ImageBuildStatusWSMessage,
) *ImageBuildWorker {
	imageBuildWorker := &ImageBuildWorker{
		store:                  store,
		successfullImageBuilds: successfullImageBuilds,
	}

	return imageBuildWorker
}

func (m *ImageBuildWorker) Run() {
	for {
		select {
		case imageBuildStatus := <-m.successfullImageBuilds:
			if imageBuildStatus.Status == "success" {
				go createDeployRequest(imageBuildStatus.BuildId, m.store)
			} else if imageBuildStatus.Status != "running" {
				go handleImageBuildError(imageBuildStatus.BuildId, m.store)
			}
		}
	}
}

func handleImageBuildError(buildId string, store *store.Store) {
	event, err := store.Event(buildId)
	if err != nil {
		logrus.Error(err)
		return
	}

	var imageBuildRequest dx.ImageBuildRequest
	err = json.Unmarshal([]byte(event.Blob), &imageBuildRequest)
	if err != nil {
		logrus.Error(err)
		return
	}

	os.RemoveAll(imageBuildRequest.SourcePath)

	event.Status = model.StatusError
	event.StatusDesc = "image build failed"
	err = store.UpdateEventStatus(event.ID, event.Status, event.StatusDesc, "[]")
	if err != nil {
		logrus.Error(err)
		return
	}
}

func createDeployRequest(buildId string, store *store.Store) {
	event, err := store.Event(buildId)
	if err != nil {
		logrus.Error(err)
		return
	}

	var imageBuildRequest dx.ImageBuildRequest
	err = json.Unmarshal([]byte(event.Blob), &imageBuildRequest)
	if err != nil {
		logrus.Error(err)
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         imageBuildRequest.Env,
		App:         imageBuildRequest.App,
		ArtifactID:  imageBuildRequest.ArtifactID,
		TriggeredBy: imageBuildRequest.TriggeredBy,
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	triggeredDeployRequestEvent, err := store.CreateEvent(&model.Event{
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: event.Repository,
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	event.Status = model.Success.String()
	event.Results = append(event.Results, model.Result{
		TriggeredDeployRequestID: triggeredDeployRequestEvent.ID,
	})
	resultsString, err := json.Marshal(event.Results)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = store.UpdateEventStatus(event.ID, event.Status, event.StatusDesc, string(resultsString))
	if err != nil {
		logrus.Error(err)
		return
	}

	os.RemoveAll(imageBuildRequest.SourcePath)
}
