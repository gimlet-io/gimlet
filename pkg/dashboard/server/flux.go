package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	fluxEvents "github.com/fluxcd/pkg/runtime/events"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func fluxEvent(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	bodyBuff := new(bytes.Buffer)
	bodyBuff.ReadFrom(rdr1)
	newStr := bodyBuff.String()
	fmt.Println(newStr)

	r.Body = rdr2 // OK since rdr2 implements the io.ReadCloser interface

	var event fluxEvents.Event
	json.NewDecoder(r.Body).Decode(&event)
	env := r.URL.Query().Get("env")

	gitopsCommit, err := asGitopsCommit(event, env)
	if err != nil {
		log.Errorf("could not translate to gitops commit: %s", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}

	ctx := r.Context()
	store := r.Context().Value("store").(*store.Store)
	repoName, repoPerEnv, err := gitopsRepoForEnv(store, env)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	stateUpdated, err := store.SaveOrUpdateGitopsCommit(gitopsCommit)
	if err != nil {
		log.Errorf("could not save or update gitops commit: %s", err)
	}

	gitopsRepoCache := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	perf := ctx.Value("perf").(*prometheus.HistogramVec)
	err = updateGitopsCommitStatuses(gitopsRepoCache, perf, store, repoName, env, event.Message, repoPerEnv)
	if err != nil {
		log.Errorf("cannot update releases: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !stateUpdated {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		return
	}

	notificationsManager := ctx.Value("notificationsManager").(notifications.Manager)
	notificationsManager.Broadcast(notifications.NewMessage(repoName, gitopsCommit, env))

	clientHub, _ := ctx.Value("clientHub").(*streaming.ClientHub)
	streaming.BroadcastGitopsCommitEvent(clientHub, *gitopsCommit)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func asGitopsCommit(event fluxEvents.Event, env string) (*model.GitopsCommit, error) {
	if _, ok := event.Metadata["revision"]; !ok {
		return nil, fmt.Errorf("could not extract gitops sha from Flux message: %s", event)
	}
	sha, err := parseRev(event.Metadata["revision"])
	if err != nil {
		return nil, err
	}

	statusDesc := event.Message

	return &model.GitopsCommit{
		Sha:        sha,
		Status:     event.Reason,
		StatusDesc: statusDesc,
		Created:    event.Timestamp.Unix(),
		Env:        env,
	}, nil
}

func parseRev(rev string) (string, error) {
	parts := strings.Split(rev, "/")
	if len(parts) != 2 {
		parts = strings.Split(rev, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("could not parse revision: %s", rev)
		}
	}

	return parts[1], nil
}

func updateGitopsCommitStatuses(
	gitopsRepoCache *nativeGit.RepoCache,
	perf *prometheus.HistogramVec,
	store *store.Store,
	repoName, env, eventMessage string,
	repoPerEnv bool,
) error {
	since := time.Now().Add(-24 * time.Hour)
	repo, pathToCleanUp, err := gitopsRepoCache.InstanceForWrite(repoName) // using a copy of the repo to avoid concurrent map writes error
	defer gitopsRepoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		return err
	}

	releases, err := gitops.Releases(repo, "", env, repoPerEnv, &since, nil, 10, "", perf)
	if err != nil {
		return err
	}

	for _, r := range releases {
		gitopsCommit, err := store.GitopsCommit(r.GitopsRef)
		if err != nil {
			log.Warnf("cannot get gitops commit: %s", err)
			continue
		}
		if gitopsCommit == nil || gitopsCommit.Status == dx.NotReconciled {
			_, err := store.SaveOrUpdateGitopsCommit(&model.GitopsCommit{
				Sha:        r.GitopsRef,
				Status:     dx.ReconciliationSucceeded,
				StatusDesc: eventMessage,
				Created:    r.Created,
				Env:        env,
			})
			if err != nil {
				log.Warnf("could not save or update gitops commit: %s", err)
			}
		}
	}
	return nil
}
