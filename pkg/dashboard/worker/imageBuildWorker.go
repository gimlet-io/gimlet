package worker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/sirupsen/logrus"
)

type ImageBuildWorker struct {
	store       *store.Store
	imageBuilds chan streaming.ImageBuildStatusWSMessage
	gitopsQueue chan int
}

func NewImageBuildWorker(
	store *store.Store,
	imageBuilds chan streaming.ImageBuildStatusWSMessage,
	gitopsQueue chan int,
) *ImageBuildWorker {
	imageBuildWorker := &ImageBuildWorker{
		store:       store,
		imageBuilds: imageBuilds,
		gitopsQueue: gitopsQueue,
	}

	return imageBuildWorker
}

func (m *ImageBuildWorker) Run() {
	for {
		imageBuildStatus := <-m.imageBuilds
		err := saveLogLine(imageBuildStatus, m.store)
		if err != nil {
			logrus.Errorf("could not save log line: %v", err)
		}

		if imageBuildStatus.Status == "success" {
			go createDeployRequest(imageBuildStatus.BuildId, m.store, m.gitopsQueue)
		} else if imageBuildStatus.Status != "running" {
			go handleImageBuildError(imageBuildStatus.BuildId, m.store)
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

func createDeployRequest(buildId string, store *store.Store, gitopsQueue chan int) {
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
		SHA:        event.SHA,
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	gitopsQueue <- 1

	event.Status = model.Success.String()
	event.Results[0].TriggeredDeployRequestID = triggeredDeployRequestEvent.ID
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

func saveLogLine(imageBuildStatus streaming.ImageBuildStatusWSMessage, dao *store.Store) error {
	event, err := dao.Event(imageBuildStatus.BuildId)
	if err != nil {
		return fmt.Errorf("could not find build with id %s", imageBuildStatus.BuildId)
	}
	if len(event.Results) == 0 {
		return nil
	}
	event.Results[0].Log += imageBuildStatus.LogLine
	resultsString, err := json.Marshal(event.Results)
	if err != nil {
		return err
	}
	err = dao.UpdateEventStatus(imageBuildStatus.BuildId, event.Status, event.StatusDesc, string(resultsString))
	if err != nil {
		return err
	}

	return nil
}
