package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Warnf("could not load .env file, relying on env vars")
	}

	config, err := config.Environ()
	if err != nil {
		logger := logrus.WithError(err)
		logger.Fatalln("main: invalid configuration")
	}

	initLogging(config)

	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		fmt.Println(config.String())
	}

	store := store.New(config.Database.Driver, config.Database.Config)

	err = setupAdminUser(config, store)
	if err != nil {
		panic(err)
	}

	var tokenManager customScm.NonImpersonatedTokenManager
	if config.Github.AppID != "" {
		tokenManager, err = customGithub.NewGithubOrgTokenManager(config)
		if err != nil {
			panic(err)
		}
	} else {
		logrus.Warnf("Please set Github Application based access for features like deleted branch detection and commit status pushing")
	}

	notificationsManager := notifications.NewManager()
	if config.Notifications.Provider == "slack" {
		notificationsManager.AddProvider(slackNotificationProvider(config))
	}
	if config.Notifications.Provider == "discord" {
		notificationsManager.AddProvider(discordNotificationProvider(config))
	}
	if tokenManager != nil {
		notificationsManager.AddProvider(notifications.NewGithubProvider(tokenManager))
	}
	go notificationsManager.Run()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	waitCh := make(chan struct{})

	parsedGitopsRepos, err := parseGitopsRepos(config.GitopsRepos)
	if err != nil {
		logrus.Warnf("could not parse gitops repositories")
	}

	repoCache, err := nativeGit.NewGitopsRepoCache(
		config.RepoCachePath,
		config.GitopsRepo,
		config.GitopsRepos,
		parsedGitopsRepos,
		config.GitopsRepoDeployKeyPath,
		stopCh,
		waitCh,
	)
	if err != nil {
		panic(err)
	}
	go repoCache.Run()
	logrus.Info("repo cache initialized")

	eventSinkHub := streaming.NewEventSinkHub(config)
	go eventSinkHub.Run()

	if config.GitopsRepo != "" &&
		config.GitopsRepoDeployKeyPath != "" {
		gitopsWorker := worker.NewGitopsWorker(
			store,
			config.GitopsRepo,
			config.GitopsRepos,
			parsedGitopsRepos,
			config.GitopsRepoDeployKeyPath,
			tokenManager,
			notificationsManager,
			eventsProcessed,
			repoCache,
			eventSinkHub,
		)
		go gitopsWorker.Run()
		logrus.Info("Gitops worker started")
	} else {
		logrus.Warn("Not starting GitOps worker. GITOPS_REPO and GITOPS_REPO_DEPLOY_KEY_PATH must be set to start GitOps worker")
	}

	if config.ReleaseStats == "enabled" {
		releaseStateWorker := &worker.ReleaseStateWorker{
			GitopsRepo: config.GitopsRepo,
			RepoCache:  repoCache,
			Releases:   releases,
			Perf:       perf,
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

	r := server.SetupRouter(config, store, notificationsManager, repoCache, perf, eventSinkHub)
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

// Creates an admin user and prints her access token, in case there are no users in the database
func setupAdminUser(config *config.Config, store *store.Store) error {
	admin, err := store.User("admin")

	if err == sql.ErrNoRows {
		admin := &model.User{
			Login:  "admin",
			Secret: adminToken(config),
			Admin:  true,
		}
		err = store.CreateUser(admin)
		if err != nil {
			return fmt.Errorf("couldn't create user admin user %s", err)
		}
		err = printAdminToken(admin)
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("couldn't list users to create admin user %s", err)
	}

	if config.PrintAdminToken {
		err = printAdminToken(admin)
		if err != nil {
			return err
		}
	} else {
		logrus.Infof("Admin token was already printed, use the PRINT_ADMIN_TOKEN=true env var to print it again")
	}

	return nil
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

func parseGitopsRepos(gitopsReposString string) ([]*config.GitopsRepoConfig, error) {
	var gitopsRepos []*config.GitopsRepoConfig
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
		gitopsRepos = append(gitopsRepos, singleGitopsRepo)
	}

	return gitopsRepos, nil
}
