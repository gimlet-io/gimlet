package main

import (
	"encoding/base32"
	"fmt"
	"log"
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
	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	if config.ReleaseStats == "enabled" {
		releaseStateWorker := &worker.ReleaseStateWorker{
			GitopsRepo:      config.GitopsRepo,
			GitopsRepos:     parsedGitopsRepos,
			DefaultRepoName: config.GitopsRepo,
			RepoCache:       repoCache,
			Releases:        releases,
			Perf:            perf,
		}
		go releaseStateWorker.Run()
	}

	if tokenManager != nil {
		branchDeleteEventWorker := worker.NewBranchDeleteEventWorker(
			tokenManager,
			config.RepoCachePath,
			store,
		)
		go branchDeleteEventWorker.Run()
	}

	metricsRouter := chi.NewRouter()
	metricsRouter.Get("/metrics", promhttp.Handler().ServeHTTP)
	go http.ListenAndServe(":8889", metricsRouter)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r := server.SetupRouter(config, store, notificationsManager, parsedGitopsRepos, repoCache, perf, eventSinkHub)
	go func() {
		err = http.ListenAndServe(":8888", r)
		if err != nil {
			panic(err)
		}
	}()

	<-waitCh
	logrus.Info("Successfully cleaned up resources. Stopping.")
}

func slackNotificationProvider(config *config.Config) *notifications.SlackProvider {
	slackChannelMap := parseChannelMap(config)

	return &notifications.SlackProvider{
		Token:          config.Notifications.Token,
		ChannelMapping: slackChannelMap,
		DefaultChannel: config.Notifications.DefaultChannel,
	}
}

func discordNotificationProvider(config *config.Config) *notifications.DiscordProvider {
	discordChannelMapping := parseChannelMap(config)

	return &notifications.DiscordProvider{
		Token:          config.Notifications.Token,
		ChannelMapping: discordChannelMapping,
		ChannelID:      config.Notifications.DefaultChannel,
	}
}

func parseChannelMap(config *config.Config) map[string]string {
	channelMap := map[string]string{}
	if config.Notifications.ChannelMapping != "" {
		pairs := strings.Split(config.Notifications.ChannelMapping, ",")
		for _, p := range pairs {
			keyValue := strings.Split(p, "=")
			channelMap[keyValue[0]] = keyValue[1]
		}
	}
	return channelMap
}

// helper function configures the logging.
func initLogging(c *config.Config) {
	if c.Logging.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if c.Logging.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
	if c.Logging.Text {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:   c.Logging.Color,
			DisableColors: !c.Logging.Color,
		})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: c.Logging.Pretty,
		})
	}
}

func printAdminToken(admin *model.User) error {
	token := token.New(token.UserToken, admin.Login)
	tokenStr, err := token.Sign(admin.Secret)
	if err != nil {
		return fmt.Errorf("couldn't create admin token %s", err)
	}
	logrus.Infof("Admin token: %s", tokenStr)

	return nil
}

func adminToken(config *config.Config) string {
	if config.AdminToken == "" {
		return base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)
	} else {
		return config.AdminToken
	}
}

func parseGitopsRepos(gitopsReposString string) (map[string]*config.GitopsRepoConfig, error) {
	gitopsRepos := map[string]*config.GitopsRepoConfig{}
	splitGitopsRepos := strings.Split(gitopsReposString, ";")

	for _, gitopsReposString := range splitGitopsRepos {
		if gitopsReposString == "" {
			continue
		}
		parsedGitopsReposString, err := url.ParseQuery(gitopsReposString)
		if err != nil {
			return nil, fmt.Errorf("invalid gitopsRepos format: %s", err)
		}
		repoPerEnv, err := strconv.ParseBool(parsedGitopsReposString.Get("repoPerEnv"))
		if err != nil {
			return nil, fmt.Errorf("invalid gitopsRepos format: %s", err)
		}

		singleGitopsRepo := &config.GitopsRepoConfig{
			Env:           parsedGitopsReposString.Get("env"),
			RepoPerEnv:    repoPerEnv,
			GitopsRepo:    parsedGitopsReposString.Get("gitopsRepo"),
			DeployKeyPath: parsedGitopsReposString.Get("deployKeyPath"),
		}
		gitopsRepos[singleGitopsRepo.Env] = singleGitopsRepo
	}

	return gitopsRepos, nil
}
