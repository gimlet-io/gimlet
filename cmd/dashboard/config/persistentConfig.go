package config

import (
	"reflect"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
)

type PersistentConfig struct {
	dao *store.Store
}

func NewPersistentConfig(dao *store.Store, config *Config) (*PersistentConfig, error) {
	p := &PersistentConfig{
		dao: dao,
	}
	err := p.Save(config)
	return p, err
}

func (p *PersistentConfig) Get(config string) (string, error) {
	return p.dao.GetConfigValue(config)
}

func (p *PersistentConfig) Save(config *Config) error {
	configs := map[string]string{}
	if config.IsGithub() {
		githubConfig := config.Github
		t := reflect.TypeOf(githubConfig)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Type.String() != "bool" {
				tag := field.Tag.Get("envconfig")
				configs[tag] = reflect.ValueOf(githubConfig).Field(i).String()
			}
		}

	} else if config.IsGitlab() {
		gitlabConfig := config.Gitlab
		t := reflect.TypeOf(gitlabConfig)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Type.String() != "bool" {
				tag := field.Tag.Get("envconfig")
				configs[tag] = reflect.ValueOf(gitlabConfig).Field(i).String()
			}
		}
	}

	for key, value := range configs {
		if value == "" {
			continue
		}

		err := p.dao.SaveConfig(&model.Config{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
