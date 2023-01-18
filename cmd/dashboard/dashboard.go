package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/cenkalti/backoff"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
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

	err = bootstrapEnvs(config.BootstrapEnv, store)
	if err != nil {
		panic(err)
	}

	notificationsManager := notifications.NewManager()
	if config.Notifications.Provider == "slack" {
		notificationsManager.AddProvider(slackNotificationProvider(config))
	}
	if config.Notifications.Provider == "discord" {
		notificationsManager.AddProvider(discordNotificationProvider(config))
	}
	go notificationsManager.Run()

	podStateManager := server.NewPodStateManager(notificationsManager, *store, 2)
	go podStateManager.Run()

	goScm := genericScm.NewGoScmHelper(config, nil)

	var gitSvc customScm.CustomGitService
	var tokenManager customScm.NonImpersonatedTokenManager

	if config.IsGithub() {
		gitSvc = &customGithub.GithubClient{}
		tokenManager, err = customGithub.NewGithubOrgTokenManager(
			config.Github.AppID,
			config.Github.InstallationID,
			config.Github.PrivateKey.String(),
		)
		if err != nil {
			panic(err)
		}
	} else if config.IsGitlab() {
		gitSvc = &customGitlab.GitlabClient{
			BaseURL: config.ScmURL(),
		}
		tokenManager = customGitlab.NewGitlabTokenManager(config.Gitlab.AdminToken)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	repoCache, err := nativeGit.NewRepoCache(
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
	go repoCache.Run()
	log.Info("repo cache initialized")

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
		repoCache,
		podStateManager,
	)
	err = http.ListenAndServe(":9000", r)
	log.Error(err)
}

// helper function configures the logging.
func initLogger(c *config.Config) {
	log.SetReportCaller(true)

	customFormatter := &log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
		},
	}
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	if c.Logging.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if c.Logging.Trace {
		log.SetLevel(log.TraceLevel)
	}
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
