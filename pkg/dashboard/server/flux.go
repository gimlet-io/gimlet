package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	fluxEvents "github.com/fluxcd/pkg/runtime/events"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
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

	notificationsManager := ctx.Value("notificationsManager").(notifications.Manager)
	notificationsManager.Broadcast(notifications.NewMessage(repoName, gitopsCommit, env))

	clientHub, _ := ctx.Value("clientHub").(*streaming.ClientHub)
	streaming.BroadcastGitopsCommitEvent(clientHub, *gitopsCommit)

	err = store.SaveOrUpdateGitopsCommit(gitopsCommit)
	if err != nil {
		log.Errorf("could not save or update gitops commit: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func asGitopsCommit(event fluxEvents.Event, env string) (*model.GitopsCommit, error) {
	if _, ok := event.Metadata["revision"]; !ok {
		return nil, fmt.Errorf("could not extract gitops sha from Flux message: %s", event)
	}
	sha := parseRev(event.Metadata["revision"])

	statusDesc := event.Message

	return &model.GitopsCommit{
		Sha:        sha,
		Status:     event.Reason,
		StatusDesc: statusDesc,
		Created:    event.Timestamp.Unix(),
		Env:        env,
	}, nil
}

func parseRev(rev string) string {
	parts := strings.Split(rev, "/")
	if len(parts) != 2 {
		return "n/a"
	}

	return parts[1]
}
