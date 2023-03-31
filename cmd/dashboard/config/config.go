package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Environ returns the settings from the environment.
func Environ() (*Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	defaults(&cfg)

	return &cfg, err
}

func DefaultChart() Chart {
	cfg := Config{}
	defaults(&cfg)
	return cfg.Chart
}

func defaults(c *Config) {
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite3"
	}
	if c.Database.Config == "" {
		c.Database.Config = "gimlet-dashboard.sqlite"
	}
	if c.RepoCachePath == "" {
		c.RepoCachePath = "/tmp/gimlet-dashboard"
	}
	if c.ReleaseHistorySinceDays == 0 {
		c.ReleaseHistorySinceDays = 30
	}
	if c.Chart.Name == "" {
		c.Chart.Name = "onechart"
	}
	if c.Chart.Repo == "" {
		c.Chart.Repo = "https://chart.onechart.dev"
	}
	if c.Chart.Version == "" {
		c.Chart.Version = "0.41.0"
	}
	if c.GitSSHAddressFormat == "" {
		c.GitSSHAddressFormat = "git@github.com:%s.git"
	}
	if c.ReleaseStats == "" {
		c.ReleaseStats = "disabled"
	}
	if c.AlertEvaluationFrequencySeconds == 0 {
		c.AlertEvaluationFrequencySeconds = 10
	}
}

// String returns the configuration in string format.
func (c *Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}

type Config struct {
	Logging                 Logging
	Host                    string `envconfig:"HOST"`
	JWTSecret               string `envconfig:"JWT_SECRET"`
	Github                  Github
	Gitlab                  Gitlab
	Database                Database
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

	TermsOfServiceFeatureFlag       bool `envconfig:"FEATURE_TERMS_OF_SERVICE"`
	AlertEvaluationFrequencySeconds int  `envconfig:"ALERT_EVALUATION_FREQUENCY_SECONDS"`
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
