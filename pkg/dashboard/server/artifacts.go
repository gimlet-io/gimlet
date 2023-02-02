package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func saveArtifact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	var artifact dx.Artifact
	json.NewDecoder(r.Body).Decode(&artifact)
	artifact.ID = fmt.Sprintf("%s-%s", artifact.Version.RepositoryName, uuid.New().String())
	artifact.Created = time.Now().Unix()

	event, err := model.ToEvent(artifact)
	if err != nil {
		logrus.Errorf("cannot convert to artifact model: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	savedEvent, err := store.CreateEvent(event)
	if err != nil {
		logrus.Errorf("cannot save artifact: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	savedArtifact, err := model.ToArtifact(savedEvent)
	artifactStr, err := json.Marshal(savedArtifact)
	if err != nil {
		logrus.Errorf("cannot serialize artifact: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(artifactStr)
}

func getArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	var limit, offset int
	var since, until *time.Time

	var repo, branch string
	var event *dx.GitEvent
	var sourceBranch string
	var sha []string

	params := r.URL.Query()
	if val, ok := params["limit"]; ok {
		l, err := strconv.Atoi(val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		limit = l
	}
	if val, ok := params["offset"]; ok {
		o, err := strconv.Atoi(val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		offset = o
	}

	if val, ok := params["since"]; ok {
		t, err := time.Parse(time.RFC3339, val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		since = &t
	}
	if val, ok := params["until"]; ok {
		t, err := time.Parse(time.RFC3339, val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		until = &t
	}

	if val, ok := params["repository"]; ok {
		repo = val[0]
	}
	if val, ok := params["branch"]; ok {
		branch = val[0]
	}
	if val, ok := params["sourceBranch"]; ok {
		sourceBranch = val[0]
	}
	if val, ok := params["sha"]; ok {
		sha = val
	}
	if val, ok := params["event"]; ok {
		event = dx.PushPtr()
		err := event.UnmarshalJSON([]byte(`"` + val[0] + `"`))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	events, err := store.Artifacts(
		repo, branch,
		event,
		sourceBranch,
		sha,
		limit, offset, since, until)
	if err != nil {
		logrus.Errorf("cannot get artifacts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	artifacts := []*dx.Artifact{}
	for _, a := range events {
		artifact, err := model.ToArtifact(a)
		if err != nil {
			logrus.Errorf("cannot deserialize artifact: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		artifacts = append(artifacts, artifact)
	}

	artifactsStr, err := json.Marshal(artifacts)
	if err != nil {
		logrus.Errorf("cannot serialize artifacts: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(artifactsStr)
}
