package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
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

	artifact, err := createDummyArtifact(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha,
		builtInEnv.Name,
		store,
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
	store *store.Store) (*dx.Artifact, error) {
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
						"repository": "nginx",
						"tag":        "latest",
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
