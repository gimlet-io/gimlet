package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
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
	gitopsRepos := ctx.Value("gitopsRepos").(map[string]*config.GitopsRepoConfig)

	repoName, repoPerEnv, err := repoInfo(gitopsRepos, env, gitopsRepo)
	if err != nil {
		logrus.Errorf("could not find repository in GITOPS_REPOS for %s and GITOPS_REPO did not provide a default repository", env)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	repo, pathToCleanUp, _, err := gitopsRepoCache.InstanceForWrite(repoName) // using a copy of the repo to avoid concurrent map writes error
	defer gitopsRepoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	perf := ctx.Value("perf").(*prometheus.HistogramVec)
	releases, err := nativeGit.Releases(repo, app, env, repoPerEnv, since, until, limit, gitRepo, perf)
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
	gitopsRepos := ctx.Value("gitopsRepos").(map[string]*config.GitopsRepoConfig)

	repoName, repoPerEnv, err := repoInfo(gitopsRepos, env, gitopsRepo)
	if err != nil {
		logrus.Errorf("could not find repository in GITOPS_REPOS for %s and GITOPS_REPO did not provide a default repository", env)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	repo := gitopsRepoCache.InstanceForRead(repoName)
	appReleases, err := nativeGit.Status(repo, app, env, repoPerEnv, perf)
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
		Type:         model.ReleaseRequestedEvent,
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
	gitopsRepoCache := ctx.Value("gitopsRepoCache").(*nativeGit.GitopsRepoCache)
	gitopsRepo := ctx.Value("gitopsRepo").(string)
	gitopsRepos := ctx.Value("gitopsRepos").(map[string]*config.GitopsRepoConfig)

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

	repoName, repoPerEnv, err := repoInfo(gitopsRepos, env, gitopsRepo)
	if err != nil {
		logrus.Errorf("could not find repository in GITOPS_REPOS for %s and GITOPS_REPO did not provide a default repository", env)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	repo, pathToCleanUp, deployKeyPath, err := gitopsRepoCache.InstanceForWrite(repoName)
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

	t0 := time.Now().UnixNano()
	head, _ := repo.Head()
	err = nativeGit.NativePush(pathToCleanUp, deployKeyPath, head.Name().Short())
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

	results := []dx.Result{}
	for _, result := range event.Results {
		results = append(results, dx.Result{
			App:                result.Manifest.App,
			Hash:               result.GitopsRef,
			Status:             result.Status.String(),
			GitopsCommitStatus: gitopsCommitStatusFromHash(store, result.GitopsRef),
			Env:                result.Manifest.Env,
			StatusDesc:         result.StatusDesc,
		})
	}

	statusBytes, _ := json.Marshal(dx.ReleaseStatus{
		Status:       event.Status,
		StatusDesc:   event.StatusDesc,
		GitopsHashes: gitopsStatus,
		Results:      results,
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

	results := []dx.Result{}
	for _, result := range event.Results {
		results = append(results, dx.Result{
			App:                result.Manifest.App,
			Hash:               result.GitopsRef,
			Status:             result.Status.String(),
			GitopsCommitStatus: gitopsCommitStatusFromHash(store, result.GitopsRef),
			Env:                result.Manifest.Env,
			StatusDesc:         result.StatusDesc,
		})
	}

	statusBytes, _ := json.Marshal(dx.ReleaseStatus{
		Status:       event.Status,
		StatusDesc:   event.StatusDesc,
		GitopsHashes: gitopsStatus,
		Results:      results,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(statusBytes)
}

func repoInfo(parsedGitopsRepos map[string]*config.GitopsRepoConfig, env string, defaultGitopsRepo string) (string, bool, error) {
	repoName := defaultGitopsRepo
	repoPerEnv := false

	if repoConfig, ok := parsedGitopsRepos[env]; ok {
		repoName = repoConfig.GitopsRepo
		repoPerEnv = repoConfig.RepoPerEnv
	}

	if repoName == "" {
		return "", false, errors.Errorf("could not find repository for %s environment and GITOPS_REPO did not provide a default repository", env)
	}

	return repoName, repoPerEnv, nil
}

func gitopsCommitStatusFromHash(store *store.Store, gitopsRef string) string {
	gitopsCommit, err := store.GitopsCommit(gitopsRef)
	if err != nil {
		logrus.Warnf("cannot get gitops commit: %s", err)
	}

	return gitopsCommit.Status
}
