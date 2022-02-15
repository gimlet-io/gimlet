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
		c.Database.Config = "gimletd.sqlite"
	}
	if c.RepoCachePath == "" {
		c.RepoCachePath = "/tmp/gimletd"
	}
	if c.ReleaseStats == "" {
		c.ReleaseStats = "disabled"
	}
}

// String returns the configuration in string format.
func (c *Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}

type Config struct {
	Debug                   bool `envconfig:"DEBUG"`
	Logging                 Logging
	Host                    string `envconfig:"HOST"`
	Database                Database
	GitopsRepo              string `envconfig:"GITOPS_REPO"`
	GitopsRepoDeployKeyPath string `envconfig:"GITOPS_REPO_DEPLOY_KEY_PATH"`
	RepoCachePath           string `envconfig:"REPO_CACHE_PATH"`
	Notifications           Notifications
	Github                  Github
	ReleaseStats            string `envconfig:"RELEASE_STATS"`
	PrintAdminToken         bool   `envconfig:"PRINT_ADMIN_TOKEN"`
}

type Database struct {
	Driver string `envconfig:"DATABASE_DRIVER"`
	Config string `envconfig:"DATABASE_CONFIG"`
}

// Logging provides the logging configuration.
type Logging struct {
	Debug  bool `envconfig:"DEBUG"`
	Trace  bool `envconfig:"TRACE"`
	Color  bool `envconfig:"LOGS_COLOR"`
	Pretty bool `envconfig:"LOGS_PRETTY"`
	Text   bool `envconfig:"LOGS_TEXT"`
}

type Notifications struct {
	Provider       string `envconfig:"NOTIFICATIONS_PROVIDER"`
	Token          string `envconfig:"NOTIFICATIONS_TOKEN"`
	DefaultChannel string `envconfig:"NOTIFICATIONS_DEFAULT_CHANNEL"`
	ChannelMapping string `envconfig:"NOTIFICATIONS_CHANNEL_MAPPING"`
}

type Github struct {
	AppID          string    `envconfig:"GITHUB_APP_ID"`
	InstallationID string    `envconfig:"GITHUB_INSTALLATION_ID"`
	PrivateKey     Multiline `envconfig:"GITHUB_PRIVATE_KEY"`
	SkipVerify     bool      `envconfig:"GITHUB_SKIP_VERIFY"`
	Debug          bool      `envconfig:"GITHUB_DEBUG"`
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
