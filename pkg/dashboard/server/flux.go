package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	fluxEvents "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	repoName, _, err := gitopsRepoForEnv(store, env)
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
	clientHub, _ := ctx.Value("clientHub").(*streaming.ClientHub)
	err = updateGitopsCommitStatuses(clientHub, gitopsRepoCache, store, event, repoName, env)
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
	clientHub *streaming.ClientHub,
	gitopsRepoCache *nativeGit.RepoCache,
	store *store.Store,
	event fluxEvents.Event,
	repoName, env string,
) error {
	if _, ok := event.Metadata["revision"]; !ok {
		return fmt.Errorf("could not extract gitops sha from Flux message: %s", event)
	}
	eventHash, err := parseRev(event.Metadata["revision"])
	if err != nil {
		return err
	}

	var commitWalker object.CommitIter
	hash := plumbing.NewHash(eventHash)
	gitopsRepoCache.PerformAction(repoName, func(repo *git.Repository) {
		commitWalker, err = repo.Log(&git.LogOptions{
			From: hash,
		})
	})
	if err != nil {
		return err
	}

	err = commitWalker.ForEach(func(c *object.Commit) error {
		hashString := c.Hash.String()
		if hashString == eventHash {
			return nil
		}

		gitopsCommitFromDb, err := store.GitopsCommit(hashString)
		if err != nil {
			log.Warnf("cannot get gitops commit: %s", err)
			return nil
		}

		if gitopsCommitFromDb != nil &&
			(gitopsCommitFromDb.Status == dx.ValidationFailed ||
				gitopsCommitFromDb.Status == dx.ReconciliationFailed ||
				gitopsCommitFromDb.Status == dx.HealthCheckFailed ||
				gitopsCommitFromDb.Status == dx.ReconciliationSucceeded) {
			return fmt.Errorf("%s", "EOF")
		}

		gitopsCommitToSave := model.GitopsCommit{
			Sha:        hashString,
			Status:     event.Reason,
			StatusDesc: event.Message,
			Created:    event.Timestamp.Unix(),
			Env:        env,
		}

		_, err = store.SaveOrUpdateGitopsCommit(&gitopsCommitToSave)
		if err != nil {
			log.Warnf("could not save or update gitops commit: %s", err)
			return nil
		}
		streaming.BroadcastGitopsCommitEvent(clientHub, gitopsCommitToSave)

		return nil
	})
	if err != nil &&
		err.Error() != "EOF" {
		return err
	}

	return nil
}
