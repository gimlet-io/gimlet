package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func getReleases(w http.ResponseWriter, r *http.Request) {
	var since, until *time.Time
	var app, env, gitRepo string
	var reverse bool
	limit := 10
	ctx := r.Context()

	params := r.URL.Query()
	if val, ok := params["limit"]; ok {
		l, err := strconv.Atoi(val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		limit = l
	}

	if val, ok := params["reverse"]; ok {
		r, err := strconv.ParseBool(val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		reverse = r
	}

	if val, ok := params["since"]; ok {
		t, err := time.Parse(time.RFC3339, val[0])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest)+" - "+err.Error(), http.StatusBadRequest)
			return
		}
		since = &t
	}
	if since == nil {
		// limiting query scope
		// without these, for apps released just once, the whole history would be traversed
		config := ctx.Value("config").(*config.Config)
		t := time.Now().Add(-1 * time.Hour * 24 * time.Duration(config.ReleaseHistorySinceDays))
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

	gitopsRepoCache := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	store := r.Context().Value("store").(*store.Store)
	repoName, repoPerEnv, err := gitopsRepoForEnv(store, env)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if repoName == "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	repo, pathToCleanUp, err := gitopsRepoCache.InstanceForWriteWithHistory(repoName) // using a copy of the repo to avoid concurrent map writes error
	defer gitopsRepoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	perf := ctx.Value("perf").(*prometheus.HistogramVec)
	releases, err := gitops.Releases(repo, app, env, repoPerEnv, since, until, limit, gitRepo, perf)
	if err != nil {
		logrus.Errorf("cannot get releases: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if reverse {
		sort.Sort(ByCreated(releases))
	}

	for _, r := range releases {
		r.GitopsRepo = repoName

		gitopsCommitStatus, gitopsCommitStatusDesc, gitopsCommitCreated := gitopsCommitMetasFromHash(store, r.GitopsRef)
		r.GitopsCommitStatus = gitopsCommitStatus
		r.GitopsCommitStatusDesc = gitopsCommitStatusDesc
		r.GitopsCommitCreated = gitopsCommitCreated
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
	gitopsRepoCache := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	perf := ctx.Value("perf").(*prometheus.HistogramVec)

	db := r.Context().Value("store").(*store.Store)
	repoName, repoPerEnv, err := gitopsRepoForEnv(db, env)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	repo, err := gitopsRepoCache.InstanceForReadWithHistory(repoName)
	if err != nil {
		logrus.Errorf("cannot get repocache: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	appReleases, err := gitops.Status(repo, app, env, repoPerEnv, perf)
	if err != nil {
		logrus.Errorf("cannot get status: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, release := range appReleases {
		if release != nil {
			release.GitopsRepo = repoName
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

func gitopsRepoForEnv(db *store.Store, env string) (string, bool, error) {
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		return "", false, fmt.Errorf("cannot get environments from database: %s", err)
	}

	for _, e := range envsFromDB {
		if e.Name == env {
			return e.AppsRepo, e.RepoPerEnv, nil
		}
	}
	return "", false, fmt.Errorf("no such environment: %s", env)
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
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: artifact.Repository,
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

func performRollback(w http.ResponseWriter, r *http.Request) {
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
		Type: model.RollbackRequestedEvent,
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
	gitopsRepoCache := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

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

	store := r.Context().Value("store").(*store.Store)
	repoName, repoPerEnv, err := gitopsRepoForEnv(store, env)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	repo, pathToCleanUp, err := gitopsRepoCache.InstanceForWriteWithHistory(repoName)
	defer gitopsRepoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(env, app)
	if repoPerEnv {
		path = app
	}

	err = nativeGit.DelDir(repo, path)
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
	if err != nil {
		logrus.Errorf("could not delete: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	t0 := time.Now().UnixNano()
	head, _ := repo.Head()
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	err = nativeGit.NativePushWithToken(pathToCleanUp, repoName, token, head.Name().Short())
	if err != nil {
		logrus.Errorf("could not push: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logrus.Infof("Pushing took %d", (time.Now().UnixNano()-t0)/1000/1000)

	gitopsRepoCache.Invalidate(repoName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func getEventReleaseTrack(w http.ResponseWriter, r *http.Request) {
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
	event, err := store.EventReleaseTrack(id)
	if err == sql.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	} else if err != nil {
		logrus.Errorf("cannot get event: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	results := []dx.Result{}
	for _, result := range event.Results {
		gitopsCommitStatus, gitopsCommitStatusDesc, _ := gitopsCommitMetasFromHash(store, result.GitopsRef)
		if event.Type == "rollback" {
			results = append(results, dx.Result{
				Hash:                   result.GitopsRef,
				Status:                 result.Status.String(),
				GitopsCommitStatus:     gitopsCommitStatus,
				GitopsCommitStatusDesc: gitopsCommitStatusDesc,
				StatusDesc:             result.StatusDesc,
				App:                    result.RollbackRequest.App,
				Env:                    result.RollbackRequest.Env,
			})
			continue
		}
		results = append(results, dx.Result{
			App:                    result.Manifest.App,
			Hash:                   result.GitopsRef,
			Status:                 result.Status.String(),
			GitopsCommitStatus:     gitopsCommitStatus,
			GitopsCommitStatusDesc: gitopsCommitStatusDesc,
			Env:                    result.Manifest.Env,
			StatusDesc:             result.StatusDesc,
		})
	}

	statusBytes, _ := json.Marshal(dx.ReleaseStatus{
		Status:     event.Status,
		StatusDesc: event.StatusDesc,
		Results:    results,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(statusBytes)
}

func getEventArtifactTrack(w http.ResponseWriter, r *http.Request) {
	var id string

	params := r.URL.Query()

	if val, ok := params["artifactId"]; ok {
		id = val[0]
	} else {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "id parameter is mandatory"), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)
	event, err := store.EventArtifactTrack(id)
	if err == sql.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	} else if err != nil {
		logrus.Errorf("cannot get event: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	results := []dx.Result{}
	for _, result := range event.Results {
		gitopsCommitStatus, gitopsCommitStatusDesc, _ := gitopsCommitMetasFromHash(store, result.GitopsRef)
		results = append(results, dx.Result{
			App:                    result.Manifest.App,
			Hash:                   result.GitopsRef,
			Status:                 result.Status.String(),
			GitopsCommitStatus:     gitopsCommitStatus,
			GitopsCommitStatusDesc: gitopsCommitStatusDesc,
			Env:                    result.Manifest.Env,
			StatusDesc:             result.StatusDesc,
		})
	}

	statusBytes, _ := json.Marshal(dx.ReleaseStatus{
		Status:     event.Status,
		StatusDesc: event.StatusDesc,
		Results:    results,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(statusBytes)
}

func gitopsCommitMetasFromHash(store *store.Store, gitopsRef string) (string, string, int64) {
	if gitopsRef == "" {
		return "", "", 0
	}
	gitopsCommit, err := store.GitopsCommit(gitopsRef)
	if err != nil {
		logrus.Warnf("cannot get gitops commit: %s", err)
		return "", "", 0
	}
	if gitopsCommit == nil {
		return "", "", 0
	}

	return gitopsCommit.Status, gitopsCommit.StatusDesc, gitopsCommit.Created
}
