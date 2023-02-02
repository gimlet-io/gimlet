package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/cenkalti/backoff"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/worker"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
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

	config, err := config.Environ()
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

	agentHub := streaming.NewAgentHub(config)
	go agentHub.Run()

	clientHub := streaming.NewClientHub()
	go clientHub.Run()

	agentWSHub := streaming.NewAgentWSHub(*clientHub)
	go agentWSHub.Run()

	store := store.New(config.Database.Driver, config.Database.Config)

	err = setupAdminUser(config, store)
	if err != nil {
		panic(err)
	}

	err = bootstrapEnvs(config.BootstrapEnv, store)
	if err != nil {
		panic(err)
	}

	gitSvc, tokenManager := initTokenManager(config)
	notificationsManager := initNotifications(config, tokenManager)

	podStateManager := server.NewPodStateManager(notificationsManager, *store, 2)
	go podStateManager.Run()

	goScm := genericScm.NewGoScmHelper(config, nil)

	stopCh := make(chan struct{})
	defer close(stopCh)

	gimletdStopCh := make(chan os.Signal, 1)
	signal.Notify(gimletdStopCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	waitCh := make(chan struct{})

	if (config.GitopsRepo == "" || config.GitopsRepoDeployKeyPath == "") && config.GitopsRepos == "" {
		log.Fatal("Either GITOPS_REPO with GITOPS_REPO_DEPLOY_KEY_PATH or GITOPS_REPOS must be set")
	}

	parsedGitopsRepos, err := parseGitopsRepos(config.GitopsRepos)
	if err != nil {
		log.Fatal("could not parse gitops repositories")
	}

	gimletdRepoCache, err := gitops.NewGitopsRepoCache(
		config.RepoCachePath,
		config.GitopsRepo,
		parsedGitopsRepos,
		config.GitopsRepoDeployKeyPath,
		config.GitSSHAddressFormat,
		gimletdStopCh,
		waitCh,
	)
	if err != nil {
		panic(err)
	}
	go gimletdRepoCache.Run()

	dashboardRepoCache, err := nativeGit.NewRepoCache(
		tokenManager,
		stopCh,
		config.RepoCachePath,
		goScm,
		config,
		clientHub,
	)
	if err != nil {
		panic(err)
	}
	go dashboardRepoCache.Run()
	log.Info("repo cache initialized")

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
		gimletdRepoCache,
		nil, // eventSinkHub,
		perf,
	)
	go gitopsWorker.Run()
	log.Info("Gitops worker started")

	metricsRouter := chi.NewRouter()
	metricsRouter.Get("/metrics", promhttp.Handler().ServeHTTP)
	go http.ListenAndServe(":9001", metricsRouter)

	go gimletdCommunication(*config, clientHub)

	r := server.SetupRouter(
		config,
		agentHub,
		clientHub,
		agentWSHub,
		store,
		gitSvc,
		tokenManager,
		dashboardRepoCache,
		podStateManager,
	)
	err = http.ListenAndServe(":9000", r)
	log.Error(err)
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

func bootstrapEnvs(envString string, store *store.Store) error {
	envsInDB, err := store.GetEnvironments()
	if err != nil {
		return err
	}

	envsToBootstrap, err := parseEnvs(envString)
	if err != nil {
		return err
	}

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

func gimletdCommunication(config config.Config, clientHub *streaming.ClientHub) {
	for {
		done := make(chan bool)

		var events chan map[string]interface{}
		var err error
		operation := func() error {
			events, err = registerGimletdEventSink(config.GimletD.URL, config.GimletD.TOKEN)
			if err != nil {
				log.Errorf("could not connect to Gimletd: %s", err.Error())
				return fmt.Errorf("could not connect to Gimletd: %s", err.Error())
			}
			return nil
		}
		backoffStrategy := backoff.NewExponentialBackOff()
		err = backoff.Retry(operation, backoffStrategy)
		if err != nil {
			log.Errorf("resetting backoff: %s", err)
			continue
		}

		log.Info("Connected to Gimletd")

		go func(events chan map[string]interface{}) {
			for {
				e, more := <-events
				if more {
					log.Debugf("event received: %v", e)

					if e["type"] == "gitopsCommit" {
						jsonString, _ := json.Marshal(streaming.GitopsEvent{
							StreamingEvent: streaming.StreamingEvent{Event: streaming.GitopsCommitEventString},
							GitopsCommit:   e["gitopsCommit"],
						})
						clientHub.Broadcast <- jsonString
					}
				} else {
					log.Info("event stream closed")
					done <- true
					return
				}
			}
		}(events)

		<-done
		log.Info("Disconnected from Gimletd")
	}
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
