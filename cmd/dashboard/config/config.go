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
		c.Chart.Version = "0.38.0"
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
	GimletD                 GimletD
	Chart                   Chart
	RepoCachePath           string `envconfig:"REPO_CACHE_PATH"`
	WebhookSecret           string `envconfig:"WEBHOOK_SECRET"`
	ReleaseHistorySinceDays int    `envconfig:"RELEASE_HISTORY_SINCE_DAYS"`
	BootstrapEnv            string `envconfig:"BOOTSTRAP_ENV"`
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
	Debug        bool   `envconfig:"GITLAB_DEBUG"`
	Org          string `envconfig:"GITLAB_ORG"`
}

type Chart struct {
	Name    string `envconfig:"CHART_NAME"`
	Repo    string `envconfig:"CHART_REPO"`
	Version string `envconfig:"CHART_VERSION"`
}

type Database struct {
	Driver string `envconfig:"DATABASE_DRIVER"`
	Config string `envconfig:"DATABASE_CONFIG"`
}

type GimletD struct {
	URL   string `envconfig:"GIMLETD_URL"`
	TOKEN string `envconfig:"GIMLETD_TOKEN"`
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

type Multiline string

func (m *Multiline) Decode(value string) error {
	value = strings.ReplaceAll(value, "\\n", "\n")
	*m = Multiline(value)
	return nil
}

func (m *Multiline) String() string {
	return string(*m)
}
