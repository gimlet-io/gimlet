package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fluxcd/pkg/runtime/events"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
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

	var event events.Event
	json.NewDecoder(r.Body).Decode(&event)
	env := r.URL.Query().Get("env")

	gitopsCommit, err := asGitopsCommit(event)
	if err != nil {
		log.Errorf("could not translate to gitops commit: %s", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}

	ctx := r.Context()
	notificationsManager := ctx.Value("notificationsManager").(notifications.Manager)
	gitopsRepo := ctx.Value("gitopsRepo").(string)
	notificationsManager.Broadcast(notifications.NewMessage(gitopsRepo, gitopsCommit, env))

	eventSinkHub := ctx.Value("eventSinkHub").(*streaming.EventSinkHub)
	eventSinkHub.BoradcastEvent(gitopsCommit)

	store := ctx.Value("store").(*store.Store)
	err = store.SaveOrUpdateGitopsCommit(gitopsCommit)
	if err != nil {
		log.Errorf("could not save or update gitops commit: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func asGitopsCommit(event events.Event) (*model.GitopsCommit, error) {
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
	}, nil
}

func parseRev(rev string) string {
	parts := strings.Split(rev, "/")
	if len(parts) != 2 {
		return "n/a"
	}

	return parts[1]
}
