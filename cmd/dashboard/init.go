package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"math/rand"
	"os"
	"path"
	"regexp"
	"runtime"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
)

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

func initTokenManager(config *config.PersistentConfig) (customScm.CustomGitService, *customScm.NonImpersonatedTokenManager) {
	var gitSvc customScm.CustomGitService
	var tokenManager customScm.NonImpersonatedTokenManager

	if config.IsGithub() {
		gitSvc = &customGithub.GithubClient{}
		var err error
		tokenManager, err = customGithub.NewGithubOrgTokenManager(
			config.Get(store.GithubAppID),
			config.Get(store.GithubInstallationID),
			config.Get(store.GithubPrivateKey),
		)
		if err != nil {
			panic(err)
		}
	} else if config.IsGitlab() {
		gitSvc = &customGitlab.GitlabClient{
			BaseURL: config.ScmURL(),
		}
		tokenManager = customGitlab.NewGitlabTokenManager(config.Get(store.GitlabAdminToken))
	} else {
		gitSvc = customScm.NewDummyGitService()
		tokenManager = customScm.NewDummyTokenManager()
	}
	return gitSvc, &tokenManager
}

func initNotifications(
	config *config.Config,
	tokenManager customScm.NonImpersonatedTokenManager,
) *notifications.ManagerImpl {
	notificationsManager := notifications.NewManager()
	if config.Notifications.Provider == "slack" {
		notificationsManager.AddProvider(slackNotificationProvider(config))
	}
	if config.Notifications.Provider == "discord" {
		notificationsManager.AddProvider(discordNotificationProvider(config))
	}
	if config.IsGithub() {
		notificationsManager.AddProvider(notifications.NewGithubProvider(tokenManager))
	} else if config.IsGitlab() {
		notificationsManager.AddProvider(notifications.NewGitlabProvider(tokenManager, config.Gitlab.URL))
	}
	go notificationsManager.Run()
	return notificationsManager
}

// helper function configures the logging.
func initLogger(c *config.Config) {
	logrus.SetReportCaller(true)

	customFormatter := &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
		},
	}
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)

	if c.Logging.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if c.Logging.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
}

func reencrypt(store *store.Store, encryptionKeyNew string) error {
	if encryptionKeyNew == "" {
		return nil
	}

	users, err := store.EverytingFromUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		store.UpdateUser(u)
		if err != nil {
			return err
		}
	}

	fmt.Println("db field re-encryption is done, please replace the value of DATABASE_ENCRYPTION_KEY with the value of DATABASE_ENCRYPTION_KEY_NEW, and delete DATABASE_ENCRYPTION_KEY_NEW environment variable")
	os.Exit(0)
	return nil
}

func bootstrapBuiltInEnv(
	store *store.Store,
) {
	envsInDB, err := store.GetEnvironments()
	if err != nil {
		panic(err)
	}
	for _, env := range envsInDB {
		if env.BuiltIn {
			return
		}
	}

	randomFirstName := firstNames[rand.Intn(len(firstNames))]
	randomSecondName := secondNames[rand.Intn(len(secondNames))]
	err = store.CreateEnvironment(&model.Environment{
		Name:    fmt.Sprintf("%s-%s", randomFirstName, randomSecondName),
		BuiltIn: true,
	})
	if err != nil {
		panic(err)
	}
}

type customFormatter struct{}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := entry.Message
	entry.Message = hideAccessToken(message)

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	}
	formatted, err := formatter.Format(entry)

	return []byte(formatted), err
}

func hideAccessToken(message string) string {
	pattern := `access_token=([^& ]+)`
	r, err := regexp.Compile(pattern)
	if err != nil {
		return message
	}
	return r.ReplaceAllString(message, "access_token=***")
}
