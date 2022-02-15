package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimletd/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

func getReleases(w http.ResponseWriter, r *http.Request) {
	var since, until *time.Time
	var app, env, gitRepo string
	limit := 10

	params := r.URL.Query()
	if val, ok := params["limit"]; ok {
		l, err := strconv.Atoi(val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		limit = l
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

	if val, ok := params["app"]; ok {
		app = val[0]
	}
	if val, ok := params["env"]; ok {
		env = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "env parameter is mandatory"), http.StatusBadRequest)
		return
	}
	if val, ok := params["git-repo"]; ok {
		gitRepo = val[0]
	}

	ctx := r.Context()
	gitopsRepoCache := ctx.Value("gitopsRepoCache").(*nativeGit.GitopsRepoCache)
	gitopsRepo := ctx.Value("gitopsRepo").(string)

	repo, pathToClanUp, err := gitopsRepoCache.InstanceForWrite() // using a copy of the repo to avoid concurrent map writes error
	defer gitopsRepoCache.CleanupWrittenRepo(pathToClanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	releases, err := nativeGit.Releases(repo, app, env, since, until, limit, gitRepo)
	if err != nil {
		logrus.Errorf("cannot get releases: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, r := range releases {
		r.GitopsRepo = gitopsRepo
	}

	releasesStr, err := json.Marshal(releases)
	if err != nil {
		logrus.Errorf("cannot serialize artifacts: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(releasesStr)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	var app, env string

	params := r.URL.Query()
	if val, ok := params["app"]; ok {
		app = val[0]
	}
	if val, ok := params["env"]; ok {
		env = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "env parameter is mandatory"), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	gitopsRepoCache := ctx.Value("gitopsRepoCache").(*nativeGit.GitopsRepoCache)
	gitopsRepo := ctx.Value("gitopsRepo").(string)
	perf := ctx.Value("perf").(*prometheus.HistogramVec)

	appReleases, err := nativeGit.Status(gitopsRepoCache.InstanceForRead(), app, env, perf)
	if err != nil {
		logrus.Errorf("cannot get status: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, release := range appReleases {
		if release != nil {
			release.GitopsRepo = gitopsRepo
			//release.Created = TODO Get githelper.Releases for each app with limit 1 - could be terribly slow
		}
	}

	appReleasesString, err := json.Marshal(appReleases)
	if err != nil {
		logrus.Errorf("cannot serialize app releases: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(appReleasesString)
}

func release(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	user := ctx.Value("user").(*model.User)

	body, _ := ioutil.ReadAll(r.Body)
	var releaseRequest dx.ReleaseRequest
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&releaseRequest)
	if err != nil {
		logrus.Errorf("cannot decode release request: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if releaseRequest.Env == "" {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "env parameter is mandatory"), http.StatusBadRequest)
		return
	}

	if releaseRequest.ArtifactID == "" {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "artifact parameter is mandatory"), http.StatusBadRequest)
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         releaseRequest.Env,
		App:         releaseRequest.App,
		ArtifactID:  releaseRequest.ArtifactID,
		TriggeredBy: user.Login,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
		return
	}

	artifact, err := store.Artifact(releaseRequest.ArtifactID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), releaseRequest.ArtifactID), http.StatusNotFound)
		return
	}
	event, err := store.CreateEvent(&model.Event{
		Type:         model.TypeRelease,
		Blob:         string(releaseRequestStr),
		Repository:   artifact.Repository,
		GitopsHashes: []string{},
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

func rollback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	user := ctx.Value("user").(*model.User)

	params := r.URL.Query()
	var env, app, targetSHA string
	if val, ok := params["env"]; ok {
		env = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "env parameter is mandatory"), http.StatusBadRequest)
		return
	}
	if val, ok := params["app"]; ok {
		app = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "app parameter is mandatory"), http.StatusBadRequest)
		return
	}
	if val, ok := params["sha"]; ok {
		targetSHA = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "sha parameter is mandatory"), http.StatusBadRequest)
		return
	}

	rollbackRequestStr, err := json.Marshal(dx.RollbackRequest{
		Env:         env,
		App:         app,
		TargetSHA:   targetSHA,
		TriggeredBy: user.Login,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot serialize rollback request: %s", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
		return
	}

	event, err := store.CreateEvent(&model.Event{
		Type: model.TypeRollback,
		Blob: string(rollbackRequestStr),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - cannot save rollback request: %s", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
		return
	}

	eventIDBytes, _ := json.Marshal(map[string]string{
		"id": event.ID,
	})

	w.WriteHeader(http.StatusCreated)
	w.Write(eventIDBytes)
}

func delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	gitopsRepoCache := ctx.Value("gitopsRepoCache").(*nativeGit.GitopsRepoCache)
	gitopsRepoDeployKeyPath := ctx.Value("gitopsRepoDeployKeyPath").(string)

	params := r.URL.Query()
	var env, app string
	if val, ok := params["env"]; ok {
		env = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "env parameter is mandatory"), http.StatusBadRequest)
		return
	}
	if val, ok := params["app"]; ok {
		app = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "app parameter is mandatory"), http.StatusBadRequest)
		return
	}

	repo, pathToCleanUp, err := gitopsRepoCache.InstanceForWrite()
	defer gitopsRepoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = nativeGit.DelDir(repo, filepath.Join(env, app))
	if err != nil {
		logrus.Errorf("cannot delete release: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		logrus.Errorf("cannot determine git status: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if empty {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	gitMessage := fmt.Sprintf("[GimletD delete] %s/%s deleted by %s", env, app, user.Login)
	_, err = nativeGit.Commit(repo, gitMessage)

	t0 := time.Now().UnixNano()
	head, _ := repo.Head()
	err = nativeGit.NativePush(pathToCleanUp, gitopsRepoDeployKeyPath, head.Name().Short())
	logrus.Infof("Pushing took %d", (time.Now().UnixNano()-t0)/1000/1000)

	gitopsRepoCache.Invalidate()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func getEvent(w http.ResponseWriter, r *http.Request) {
	var id string

	params := r.URL.Query()

	if val, ok := params["id"]; ok {
		id = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "id parameter is mandatory"), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	event, err := store.Event(id)
	if err == sql.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	} else if err != nil {
		logrus.Errorf("cannot get event: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	gitopsStatus := []dx.GitopsStatus{}
	for _, gitopsHash := range event.GitopsHashes {
		gitopsCommit, err := store.GitopsCommit(gitopsHash)
		if err != nil {
			logrus.Warnf("cannot get gitops commit: %s", err)
			continue
		}

		if gitopsCommit != nil {
			gitopsStatus = append(gitopsStatus, dx.GitopsStatus{
				Hash:       gitopsHash,
				Status:     gitopsCommit.Status,
				StatusDesc: gitopsCommit.StatusDesc,
			})
		} else {
			gitopsStatus = append(gitopsStatus, dx.GitopsStatus{
				Hash:   gitopsHash,
				Status: "N/A",
			})
		}
	}

	statusBytes, _ := json.Marshal(dx.ReleaseStatus{
		Status:       event.Status,
		StatusDesc:   event.StatusDesc,
		GitopsHashes: gitopsStatus,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(statusBytes)
}
