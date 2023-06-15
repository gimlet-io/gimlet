package config

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/kelseyhightower/envconfig"
)

const CONFIG_STORAGE_KEY = "config"

const DEFAULT_CHART_NAME = "onechart"
const DEFAULT_CHART_REPO = "https://chart.onechart.dev"
const DEFAULT_CHART_VERSION = "0.47.0"

// LoadConfig persist env config in the db and loads config from the db.
// DB values take precedence
func LoadConfig(dao *store.Store) (*Config, error) {
	// First we load config from the environment
	cfg := &Config{
		dao: dao,
	}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	// Then we set defaults
	defaults(cfg)

	// Then we persist it to the database if they are not set yet
	err = cfg.persistEnvConfig()
	if err != nil {
		return nil, err
	}

	// Then we reload from the DB, so DB values take precedence
	err = cfg.load()
	return cfg, err
}

// LoadStaticConfig returns the static config from the environment.
func LoadStaticConfig() (*StaticConfig, error) {
	cfg := StaticConfig{}
	err := envconfig.Process("", &cfg)
	staticDefaults(&cfg)

	return &cfg, err
}

func staticDefaults(c *StaticConfig) {
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite3"
	}
	if c.Database.Config == "" {
		c.Database.Config = "gimlet-dashboard.sqlite"
	}
}

func defaults(c *Config) {
	if c.RepoCachePath == "" {
		c.RepoCachePath = "/tmp/gimlet-dashboard"
	}
	if c.ReleaseHistorySinceDays == 0 {
		c.ReleaseHistorySinceDays = 30
	}
	if c.Chart.Name == "" {
		c.Chart.Name = DEFAULT_CHART_NAME
	}
	if c.Chart.Repo == "" {
		c.Chart.Repo = DEFAULT_CHART_REPO
	}
	if c.Chart.Version == "" {
		c.Chart.Version = DEFAULT_CHART_VERSION
	}
	if c.GitSSHAddressFormat == "" {
		c.GitSSHAddressFormat = "git@github.com:%s.git"
	}
	if c.ReleaseStats == "" {
		c.ReleaseStats = "disabled"
	}
}

// StaticConfig holds Gimlet configuration that can only be set with environment variables
type StaticConfig struct {
	Logging  Logging
	Database Database
}

// Config holds Gimlet configuration that is stored in the database
// * It can be initiatied with environment variables
// * and dynamically changed runtime that is persisted in the database.
// * Values in the database take precedence.
//
// We have a single instance of this struct in Gimlet
// changes to this struct are reflected application wide as it has a pointer reference.
// To make the changes persistent, call Persist()
type Config struct {
	dao *store.Store

	Host      string `envconfig:"HOST"`
	JWTSecret string `envconfig:"JWT_SECRET"`
	Github    Github
	Gitlab    Gitlab

	Notifications           Notifications
	Chart                   Chart
	RepoCachePath           string `envconfig:"REPO_CACHE_PATH"`
	WebhookSecret           string `envconfig:"WEBHOOK_SECRET"`
	ReleaseHistorySinceDays int    `envconfig:"RELEASE_HISTORY_SINCE_DAYS"`
	BootstrapEnv            string `envconfig:"BOOTSTRAP_ENV"`
	UserflowToken           string `envconfig:"USERFLOW_TOKEN"`

	PrintAdminToken bool   `envconfig:"PRINT_ADMIN_TOKEN"`
	AdminToken      string `envconfig:"ADMIN_TOKEN"`

	// Deprecated, use BootstrapEnv instead
	GitopsRepo string `envconfig:"GITOPS_REPO"`
	// Deprecated, use BootstrapEnv instead
	GitopsRepos string `envconfig:"GITOPS_REPOS"`

	GitopsRepoDeployKeyPath string `envconfig:"GITOPS_REPO_DEPLOY_KEY_PATH"`
	GitSSHAddressFormat     string `envconfig:"GIT_SSH_ADDRESS_FORMAT"`
	ReleaseStats            string `envconfig:"RELEASE_STATS"`

	TermsOfServiceFeatureFlag      bool `envconfig:"FEATURE_TERMS_OF_SERVICE"`
	ChartVersionUpdaterFeatureFlag bool `envconfig:"FEATURE_CHART_VERSION_UPDATER"`
}

// persist all config fields that are not already set in the database
func (c *Config) persistEnvConfig() error {
	var persistedConfig Config
	persistedConfigString, err := c.dao.KeyValue(CONFIG_STORAGE_KEY)
	if err == nil {
		err = json.Unmarshal([]byte(persistedConfigString.Value), &persistedConfig)
		if err != nil {
			return err
		}
	} else if err == sql.ErrNoRows {
		persistedConfig = Config{}
	} else if err != nil {
		return err
	}

	updateConfigWhenZeroValue(&persistedConfig, c)
	persistedConfig.dao = c.dao

	return persistedConfig.persist()
}

func updateConfigWhenZeroValue(toUpdate *Config, new *Config) {
	t := reflect.TypeOf(*toUpdate)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := reflect.ValueOf(*toUpdate).Field(i)
		newValue := reflect.ValueOf(*new).FieldByName(field.Name)

		if value.Kind() == reflect.Struct {
			for j := 0; j < value.NumField(); j++ {
				nestedField := value.Type().Field(j)
				nestedValue := value.Field(j)
				newNestedValue := newValue.FieldByName(nestedField.Name)

				if nestedValue.IsZero() {
					obj := reflect.ValueOf(toUpdate).Elem()
					if nestedValue.Kind() == reflect.Bool {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetBool(newNestedValue.Bool())
					} else if nestedValue.Kind() == reflect.Int {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetInt(newNestedValue.Int())
					} else if nestedValue.Kind() == reflect.String {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetString(newNestedValue.String())
					}
				}
			}
		} else {
			obj := reflect.ValueOf(toUpdate).Elem()

			if value.IsZero() {
				if value.Kind() == reflect.Bool {
					obj.FieldByName(field.Name).SetBool(newValue.Bool())
				} else if value.Kind() == reflect.Int {
					obj.FieldByName(field.Name).SetInt(newValue.Int())
				} else if value.Kind() == reflect.String {
					obj.FieldByName(field.Name).SetString(newValue.String())
				}
			}
		}
	}
}

// Persist saves the config struct to the DB
// When we are updating configs, like regsistering new github applications,
// the config user should call a Persist on the config
func (c *Config) Persist() error {
	configString, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return c.dao.SaveKeyValue(&model.KeyValue{
		Key:   CONFIG_STORAGE_KEY,
		Value: string(configString),
	})
}

func (c *Config) load() error {
	configString, err := c.dao.KeyValue(CONFIG_STORAGE_KEY)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(configString.Value), c)
}

// Logging provides the logging configuration.
type Logging struct {
	Debug bool `envconfig:"DEBUG"`
	Trace bool `envconfig:"TRACE"`
}

type Github struct {
	AppID          string    `envconfig:"GITHUB_APP_ID"`
	InstallationID string    `envconfig:"GITHUB_INSTALLATION_ID"`
	PrivateKey     Multiline `envconfig:"GITHUB_PRIVATE_KEY"`
	ClientID       string    `envconfig:"GITHUB_CLIENT_ID"`
	ClientSecret   string    `envconfig:"GITHUB_CLIENT_SECRET"`
	SkipVerify     bool      `envconfig:"GITHUB_SKIP_VERIFY"`
	Debug          bool      `envconfig:"GITHUB_DEBUG"`
	Org            string    `envconfig:"GITHUB_ORG"`
}

type Gitlab struct {
	ClientID     string `envconfig:"GITLAB_CLIENT_ID"`
	ClientSecret string `envconfig:"GITLAB_CLIENT_SECRET"`
	// This is a personal access token of the Gitlab admin or a Group Token
	AdminToken string `envconfig:"GITLAB_ADMIN_TOKEN"`
	Debug      bool   `envconfig:"GITLAB_DEBUG"`
	Org        string `envconfig:"GITLAB_ORG"`
	URL        string `envconfig:"GITLAB_URL"`
}

type Chart struct {
	Name    string `envconfig:"CHART_NAME"`
	Repo    string `envconfig:"CHART_REPO"`
	Version string `envconfig:"CHART_VERSION"`
}

type Database struct {
	Driver           string `envconfig:"DATABASE_DRIVER"`
	Config           string `envconfig:"DATABASE_CONFIG"`
	EncryptionKey    string `envconfig:"DATABASE_ENCRYPTION_KEY"`
	EncryptionKeyNew string `envconfig:"DATABASE_ENCRYPTION_KEY_NEW"`
}

type Notifications struct {
	Provider       string `envconfig:"NOTIFICATIONS_PROVIDER"`
	Token          string `envconfig:"NOTIFICATIONS_TOKEN"`
	DefaultChannel string `envconfig:"NOTIFICATIONS_DEFAULT_CHANNEL"`
	ChannelMapping string `envconfig:"NOTIFICATIONS_CHANNEL_MAPPING"`
}

type GitopsRepoConfig struct {
	Env           string
	RepoPerEnv    bool
	GitopsRepo    string
	DeployKeyPath string
}

func (c *Config) IsGithub() bool {
	return c.Github.AppID != ""
}

func (c *Config) IsGitlab() bool {
	return c.Gitlab.ClientID != ""
}

func (c *Config) Org() string {
	if c.IsGithub() {
		return c.Github.Org
	} else if c.IsGitlab() {
		return c.Gitlab.Org
	}

	return ""
}

func (c *Config) ScmURL() string {
	if c.IsGithub() {
		return "https://github.com"
	} else if c.IsGitlab() {
		if c.Gitlab.URL != "" {
			return c.Gitlab.URL
		}
		return "https://gitlab.com"
	}

	return ""
}

type Multiline string

func (m *Multiline) Decode(value string) error {
	value = strings.ReplaceAll(value, "\\n", "\n")
	*m = Multiline(value)
	return nil
}

func (m *Multiline) String() string {
	return string(*m)
}
