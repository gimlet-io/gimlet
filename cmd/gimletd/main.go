package main

import (
	"encoding/base32"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strconv"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
)

func main() {

	eventSinkHub := streaming.NewEventSinkHub(config)
	go eventSinkHub.Run()

	gitopsWorker := worker.NewGitopsWorker(
		store,
		config.GitopsRepo,
		parsedGitopsRepos,
		config.GitopsRepoDeployKeyPath,
		tokenManager,
		notificationsManager,
		eventsProcessed,
		repoCache,
		eventSinkHub,
		perf,
	)
	go gitopsWorker.Run()
	logrus.Info("Gitops worker started")

	r := server.SetupRouter(config, store, notificationsManager, parsedGitopsRepos, repoCache, perf, eventSinkHub)
