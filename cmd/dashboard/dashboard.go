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

	dconfig "github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/alert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	dstore "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/worker"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Warnf("could not load .env file, relying on env vars")
	}

	config, err := dconfig.Environ()
	if err != nil {
		log.Fatalln("main: invalid configuration")
	}

	initLogger(config)
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Traceln(config.String())
	}

	if config.Host == "" {
		panic(fmt.Errorf("please provide the HOST variable"))
	}

	if config.JWTSecret == "" {
		panic(fmt.Errorf("please provide the JWT_SECRET variable"))
	}

	agentHub := streaming.NewAgentHub()
	go agentHub.Run()

	clientHub := streaming.NewClientHub()
	go clientHub.Run()

	agentWSHub := streaming.NewAgentWSHub(*clientHub)
	go agentWSHub.Run()

	store := dstore.New(
		config.Database.Driver,
		config.Database.Config,
		config.Database.EncryptionKey,
		config.Database.EncryptionKeyNew,
	)

	persistentConfig, err := dconfig.NewPersistentConfig(store, config)
	if err != nil {
		panic(err)
	}

	err = reencrypt(store, persistentConfig.Get(dstore.DatabaseEncryptionKeyNew))
	if err != nil {
		panic(err)
	}

	err = setupAdminUser(persistentConfig, store)
	if err != nil {
		panic(err)
	}

	err = bootstrapEnvs(persistentConfig.Get(dstore.BootstrapEnv), store, "")
	if err != nil {
		panic(err)
	}

	gitSvc, tokenManager := initTokenManager(persistentConfig)
	notificationsManager := initNotifications(persistentConfig, tokenManager)

	alertStateManager := alert.NewAlertStateManager(notificationsManager, *store, 2)
	// go alertStateManager.Run()

	goScm := genericScm.NewGoScmHelper(persistentConfig, nil)

	stopCh := make(chan struct{})
	defer close(stopCh)

	gimletdStopCh := make(chan os.Signal, 1)
	signal.Notify(gimletdStopCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	waitCh := make(chan struct{})

	if persistentConfig.Get(dstore.GitopsRepo) != "" || persistentConfig.Get(dstore.GitopsRepoDeployKeyPath) != "" {
		panic("GITOPS_REPO and GITOPS_REPO_DEPLOY_KEY_PATH are deprecated." +
			"Please use BOOTSTRAP_ENV instead, or create gitops environment configurations on the Gimlet dashboard.")
	}

	gitopsRepos := persistentConfig.Get(dstore.GitopsRepos)
	if gitopsRepos != "" {
		log.Info("Bootstrapping gitops environments from deprecated GITOPS_REPOS variable")
		err = bootstrapEnvs(
			persistentConfig.Get(dstore.BootstrapEnv),
			store,
			gitopsRepos,
		)
		if err != nil {
			panic(err)
		}
		log.Info("Gitops environments bootstrapped, remove the deprecated GITOPS_REPO* vars. They are now written to the Gimlet database." +
			"You can also delete the deploykey. The gitops repo is accessed via the Github Application / Gitlab admin token.")
	}

	repoCachePath := persistentConfig.Get(dstore.RepoCachePath)
	host := persistentConfig.Get(dstore.Host)
	webhookSecrets := persistentConfig.Get(dstore.WebhookSecret)
	dashboardRepoCache, err := nativeGit.NewRepoCache(
		tokenManager,
		stopCh,
		repoCachePath,
		host,
		webhookSecrets,
		goScm,
		persistentConfig,
		clientHub,
	)
	if err != nil {
		panic(err)
	}
	go dashboardRepoCache.Run()
	log.Info("repo cache initialized")

	chartUpdatePullRequests := map[string]interface{}{}
	if persistentConfig.ChartVersionUpdaterFeatureFlag() {
		chart := dconfig.Chart{
			Name:    persistentConfig.Get(dstore.ChartName),
			Version: persistentConfig.Get(dstore.ChartVersion),
		}

		chartVersionUpdater := worker.NewChartVersionUpdater(
			gitSvc,
			tokenManager,
			dashboardRepoCache,
			goScm,
			&chartUpdatePullRequests,
			chart,
		)
		go chartVersionUpdater.Run()
	}

	gitopsWorker := worker.NewGitopsWorker(
		store,
		persistentConfig.Get(dstore.GitopsRepo),
		persistentConfig.Get(dstore.GitopsRepoDeployKeyPath),
		tokenManager,
		notificationsManager,
		eventsProcessed,
		dashboardRepoCache,
		clientHub,
		perf,
	)
	go gitopsWorker.Run()
	log.Info("Gitops worker started")

	if persistentConfig.Get(dstore.ReleaseStats) == "enabled" {
		releaseStateWorker := &worker.ReleaseStateWorker{
			RepoCache: dashboardRepoCache,
			Releases:  releases,
			Perf:      perf,
			Store:     store,
			Config:    persistentConfig,
		}
		go releaseStateWorker.Run()
	}

	branchDeleteEventWorker := worker.NewBranchDeleteEventWorker(
		tokenManager,
		persistentConfig.Get(dstore.RepoCachePath),
		store,
	)
	go branchDeleteEventWorker.Run()

	metricsRouter := chi.NewRouter()
	metricsRouter.Get("/metrics", promhttp.Handler().ServeHTTP)
	go http.ListenAndServe(":9001", metricsRouter)

	logger := log.New()
	logger.Formatter = &customFormatter{}

	r := server.SetupRouter(
		config,
		persistentConfig,
		agentHub,
		clientHub,
		agentWSHub,
		store,
		gitSvc,
		tokenManager,
		dashboardRepoCache,
		&chartUpdatePullRequests,
		alertStateManager,
		notificationsManager,
		perf,
		logger,
	)

	go func() {
		err = http.ListenAndServe(":9000", r)
		if err != nil {
			panic(err)
		}
	}()

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
	store *dstore.Store,
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

func slackNotificationProvider(config *dconfig.PersistentConfig) *notifications.SlackProvider {
	slackChannelMap := parseChannelMap(config.Get(store.NotificationsChannelMapping))

	return &notifications.SlackProvider{
		Token:          config.Get(store.NotificationsToken),
		ChannelMapping: slackChannelMap,
		DefaultChannel: config.Get(store.NotificationsDefaultChannel),
	}
}

func discordNotificationProvider(config *dconfig.PersistentConfig) *notifications.DiscordProvider {
	discordChannelMapping := parseChannelMap(config.Get(store.NotificationsChannelMapping))

	return &notifications.DiscordProvider{
		Token:          config.Get(store.NotificationsToken),
		ChannelMapping: discordChannelMapping,
		ChannelID:      config.Get(store.NotificationsDefaultChannel),
	}
}

func parseChannelMap(notificationsChannelMapping string) map[string]string {
	channelMap := map[string]string{}
	if notificationsChannelMapping != "" {
		pairs := strings.Split(notificationsChannelMapping, ",")
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
