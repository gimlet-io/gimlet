package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Environ returns the settings from the environment.
func Environ() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)

	defaults(&cfg)
	return cfg, err
}

func defaults(c *Config) {
	if c.ImageBuilderHost == "" {
		c.ImageBuilderHost = "http://image-builder.infrastructure.svc.cluster.local:9000/build-image"
	}
}

// String returns the configuration in string format.
func (c *Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}

type Config struct {
	Logging          Logging
	KubeConfig       string `envconfig:"KUBECONFIG"`
	Env              string `envconfig:"ENV"`
	Namespace        string `envconfig:"NAMESPACE"`
	Host             string `envconfig:"HOST"`
	AgentKey         string `envconfig:"AGENT_KEY"`
	ImageBuilderHost string `envconfig:"IMAGE_BUILDER_HOST"`
}

// Logging provides the logging configuration.
type Logging struct {
	Debug bool `envconfig:"DEBUG"`
	Trace bool `envconfig:"TRACE"`
}
