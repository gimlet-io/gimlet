package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/alert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/worker"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println(logo())

	godotenv.Load(".env")

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("main: invalid configuration")
	}

	initLogger(config)

	store := store.New(
		config.Database.Driver,
		config.Database.Config,
		config.Database.EncryptionKey,
		config.Database.EncryptionKeyNew,
	)

	dynamicConfig, err := dynamicconfig.LoadDynamicConfig(store)
	if err != nil {
		panic(err)
	}

	log.Infof("Admin auth key: %s", adminKey(dynamicConfig))

	if config.Host == "" {
		panic(fmt.Errorf("please provide the HOST variable"))
	}

	if dynamicConfig.JWTSecret == "" {
		generateAndPersistJwtSecret(dynamicConfig)
	}

	agentHub := streaming.NewAgentHub()
	go agentHub.Run()

	clientHub := streaming.NewClientHub()
	go clientHub.Run()

	successfullImageBuilds := make(chan streaming.ImageBuildStatusWSMessage)
	agentWSHub := streaming.NewAgentWSHub(*clientHub, successfullImageBuilds)
	go agentWSHub.Run()

	err = reencrypt(store, config.Database.EncryptionKeyNew)
	if err != nil {
		panic(err)
	}

	err = setupAdminUser(config, store)
	if err != nil {
		panic(err)
	}

	err = bootstrapEnvs(config.BootstrapEnv, store, "")
	if err != nil {
		panic(err)
	}

	tokenManager := customScm.NewTokenManager(dynamicConfig)
	notificationsManager := initNotifications(config, dynamicConfig, tokenManager)

	alertStateManager := alert.NewAlertStateManager(notificationsManager, *store, 2)
	// go alertStateManager.Run()

	stopCh := make(chan struct{})
	defer close(stopCh)

	gimletdStopCh := make(chan os.Signal, 1)
	signal.Notify(gimletdStopCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	waitCh := make(chan struct{})

	if config.GitopsRepo != "" || config.GitopsRepoDeployKeyPath != "" {
		panic("GITOPS_REPO and GITOPS_REPO_DEPLOY_KEY_PATH are deprecated." +
			"Please use BOOTSTRAP_ENV instead, or create gitops environment configurations on the Gimlet dashboard.")
	}

	if config.GitopsRepos != "" {
		log.Info("Bootstrapping gitops environments from deprecated GITOPS_REPOS variable")
		err = bootstrapEnvs(
			config.BootstrapEnv,
			store,
			config.GitopsRepos,
		)
		if err != nil {
			panic(err)
		}
		log.Info("Gitops environments bootstrapped, remove the deprecated GITOPS_REPO* vars. They are now written to the Gimlet database." +
			"You can also delete the deploykey. The gitops repo is accessed via the Github Application / Gitlab admin token.")
	}

	gitUser, err := setupGitUser(config, store)
	if err != nil {
		panic(err)
	}

	repoCache, err := nativeGit.NewRepoCache(
		tokenManager,
		stopCh,
		config,
		dynamicConfig,
		clientHub,
		gitUser,
	)
	if err != nil {
		panic(err)
	}
	go repoCache.Run()
	log.Info("Repo cache initialized")

	imageBuilds := map[string]streaming.ImageBuildTrigger{}
	imageBuildWorker := worker.NewImageBuildWorker(store, successfullImageBuilds)
	go imageBuildWorker.Run()

	chartUpdatePullRequests := map[string]interface{}{}
	if config.ChartVersionUpdaterFeatureFlag {
		chartVersionUpdater := worker.NewChartVersionUpdater(
			config,
			dynamicConfig,
			tokenManager,
			repoCache,
			&chartUpdatePullRequests,
		)
		go chartVersionUpdater.Run()
	}

	gitopsWorker := worker.NewGitopsWorker(
		store,
		config.GitopsRepo,
		config.GitopsRepoDeployKeyPath,
		tokenManager,
		notificationsManager,
		eventsProcessed,
		repoCache,
		clientHub,
		perf,
		gitUser,
		config.GitHost,
	)
	go gitopsWorker.Run()
	log.Info("Gitops worker started")

	if config.ReleaseStats == "enabled" {
		releaseStateWorker := &worker.ReleaseStateWorker{
			RepoCache:     repoCache,
			Releases:      releases,
			Perf:          perf,
			Store:         store,
			DynamicConfig: dynamicConfig,
		}
		go releaseStateWorker.Run()
	}

	branchDeleteEventWorker := worker.NewBranchDeleteEventWorker(
		tokenManager,
		config.RepoCachePath,
		store,
	)
	go branchDeleteEventWorker.Run()

	metricsRouter := chi.NewRouter()
	metricsRouter.Get("/metrics", promhttp.Handler().ServeHTTP)
	go http.ListenAndServe(":9001", metricsRouter)

	logger := log.New()
	logger.Formatter = &customFormatter{}

	var gitServer http.Handler
	if config.BuiltinEnvFeatureFlag() {
		gitServer, err = builtInGitServer(gitUser, config.GitRoot)
		if err != nil {
			panic(err)
		}
	}

	r := server.SetupRouter(
		config,
		dynamicConfig,
		agentHub,
		clientHub,
		agentWSHub,
		store,
		tokenManager,
		repoCache,
		&chartUpdatePullRequests,
		alertStateManager,
		notificationsManager,
		perf,
		logger,
		gitServer,
		imageBuilds,
		gitUser,
	)

	go func() {
		err = http.ListenAndServe(":9000", r)
		if err != nil {
			panic(err)
		}
	}()

	if config.BuiltinEnvFeatureFlag() {
		time.Sleep(time.Millisecond * 100) // wait til the router is up
		err = bootstrapBuiltInEnv(store, repoCache, gitUser, config, dynamicConfig)
		if err != nil {
			panic(err)
		}
	}

	<-waitCh
	log.Info("Successfully cleaned up resources. Stopping.")
}

func parseEnvs(envString string) ([]*model.Environment, error) {
	var envs []*model.Environment
	splitEnvString := strings.Split(envString, ";")

	for _, envString := range splitEnvString {
		if envString == "" {
			continue
		}
		parsedEnvString, err := url.ParseQuery(envString)
		if err != nil {
			return nil, fmt.Errorf("invalid env format: %s", err)
		}
		repoPerEnv, err := strconv.ParseBool(parsedEnvString.Get("repoPerEnv"))
		if err != nil {
			return nil, fmt.Errorf("invalid env format: %s", err)
		}

		env := &model.Environment{
			Name:       parsedEnvString.Get("name"),
			RepoPerEnv: repoPerEnv,
			InfraRepo:  parsedEnvString.Get("infraRepo"),
			AppsRepo:   parsedEnvString.Get("appsRepo"),
		}
		envs = append(envs, env)
	}
	return envs, nil
}

func bootstrapEnvs(
	envString string,
	store *store.Store,
	gitopsRepos string,
) error {
	envsInDB, err := store.GetEnvironments()
	if err != nil {
		return err
	}

	envsToBootstrap, err := parseEnvs(envString)
	if err != nil {
		return err
	}

	deprecatedGitopsReposToEnvs, err := parseGitopsRepos(gitopsRepos)
	if err != nil {
		return err
	}
	envsToBootstrap = append(envsToBootstrap, deprecatedGitopsReposToEnvs...)

	for _, envToBootstrap := range envsToBootstrap {
		if !envExists(envsInDB, envToBootstrap.Name) {
			if envToBootstrap.Name == "" || envToBootstrap.InfraRepo == "" || envToBootstrap.AppsRepo == "" {
				return fmt.Errorf("name, infraRepo, and appsRepo are mandatory for environments")
			}
			err := store.CreateEnvironment(envToBootstrap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func envExists(envsInDB []*model.Environment, envName string) bool {
	for _, env := range envsInDB {
		if env.Name == envName {
			return true
		}
	}

	return false
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

func parseGitopsRepos(gitopsReposString string) ([]*model.Environment, error) {
	envs := []*model.Environment{}
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

		env := &model.Environment{
			Name:       parsedGitopsReposString.Get("env"),
			RepoPerEnv: repoPerEnv,
			AppsRepo:   parsedGitopsReposString.Get("gitopsRepo"),
			InfraRepo:  "migrated from Gimletd config, ask on Discord how to migrate",
		}
		envs = append(envs, env)
	}

	return envs, nil
}

func logo() string {
	return `   _____ _____ __  __ _      ______ _______
  / ____|_   _|  \/  | |    |  ____|__   __|
 | |  __  | | | \  / | |    | |__     | |
 | | |_ | | | | |\/| | |    |  __|    | |
 | |__| |_| |_| |  | | |____| |____   | |
  \_____|_____|_|  |_|______|______|  |_|
`
}
