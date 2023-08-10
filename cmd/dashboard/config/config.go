package config

import (
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

const DEFAULT_CHART_NAME = "onechart"
const DEFAULT_CHART_REPO = "https://chart.onechart.dev"
const DEFAULT_CHART_VERSION = "0.50.0"
const DEFAULT_CHARTS = "name=onechart,repo=https://chart.onechart.dev,version=0.50.0;name=static-site,repo=https://chart.onechart.dev,version=0.1.0"

// LoadConfig returns the static config from the environment.
func LoadConfig() (*Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	defaults(&cfg)

	return &cfg, err
}

func defaults(c *Config) {
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite"
	}
	if c.Database.Config == "" {
		c.Database.Config = "gimlet-dashboard.sqlite?_pragma=busy_timeout=10000"
	}
	if c.RepoCachePath == "" {
		c.RepoCachePath = "/tmp/gimlet-dashboard"
	}
	if c.ReleaseHistorySinceDays == 0 {
		c.ReleaseHistorySinceDays = 30
	}
	if c.Charts == "" {
		c.Charts = DEFAULT_CHARTS
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
	if c.GitRoot == "" {
		c.GitRoot = "git-root/"
	}
	if c.GitHost == "" {
		c.GitHost = "127.0.0.1:9000"
	}
	if c.ApiHost == "" {
		c.ApiHost = c.Host
	}
	if c.BuiltinEnvFeatureFlagString == "" {
		c.BuiltinEnvFeatureFlagString = "true"
	}
}

// Config holds Gimlet configuration that can only be set with environment variables
type Config struct {
	Logging  Logging
	Database Database

	Host      string `envconfig:"HOST"`
	JWTSecret string `envconfig:"JWT_SECRET"`
	Github    Github
	Gitlab    Gitlab

	Notifications           Notifications
	Chart                   Chart
	Charts                  string `envconfig:"CHARTS"`
	RepoCachePath           string `envconfig:"REPO_CACHE_PATH"`
	WebhookSecret           string `envconfig:"WEBHOOK_SECRET"`
	ReleaseHistorySinceDays int    `envconfig:"RELEASE_HISTORY_SINCE_DAYS"`
	BootstrapEnv            string `envconfig:"BOOTSTRAP_ENV"`

	AdminToken string `envconfig:"ADMIN_TOKEN"`

	// Deprecated, use BootstrapEnv instead
	GitopsRepo string `envconfig:"GITOPS_REPO"`
	// Deprecated, use BootstrapEnv instead
	GitopsRepos string `envconfig:"GITOPS_REPOS"`

	GitopsRepoDeployKeyPath string `envconfig:"GITOPS_REPO_DEPLOY_KEY_PATH"`
	GitSSHAddressFormat     string `envconfig:"GIT_SSH_ADDRESS_FORMAT"`
	ReleaseStats            string `envconfig:"RELEASE_STATS"`

	TermsOfServiceFeatureFlag      bool   `envconfig:"FEATURE_TERMS_OF_SERVICE"`
	ChartVersionUpdaterFeatureFlag bool   `envconfig:"FEATURE_CHART_VERSION_UPDATER"`
	BuiltinEnvFeatureFlagString    string `envconfig:"FEATURE_BUILT_IN_ENV"`

	GitHost          string `envconfig:"GIT_HOST"`
	ApiHost          string `envconfig:"API_HOST"`
	GitRoot          string `envconfig:"GIT_ROOT"`
	ImageBuilderHost string `envconfig:"IMAGE_BUILDER_HOST"`
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

type Multiline string

func (m *Multiline) Decode(value string) error {
	value = strings.ReplaceAll(value, "\\n", "\n")
	*m = Multiline(value)
	return nil
}

func (m *Multiline) String() string {
	return string(*m)
}

func (c *Config) BuiltinEnvFeatureFlag() bool {
	flag, err := strconv.ParseBool(c.BuiltinEnvFeatureFlagString)
	if err != nil {
		logrus.Warnf("could not parse FEATURE_BUILT_IN_ENV: %s", err)
		return true
	}
	return flag
}
