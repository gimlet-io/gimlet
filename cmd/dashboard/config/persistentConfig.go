package config

import (
	"reflect"
	"strconv"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

type PersistentConfig struct {
	dao *store.Store
}

func NewPersistentConfig(dao *store.Store, config *Config) (*PersistentConfig, error) {
	p := &PersistentConfig{
		dao: dao,
	}
	err := p.saveConfigFile(config)
	return p, err
}

func (p *PersistentConfig) Get(key string) string {
	config, err := p.dao.GetConfig(key)
	if err != nil {
		logrus.Warnf("cannot get config %s from db: %s", key, err)
	}
	return config.Value
}

func (p *PersistentConfig) Save(key string, value string) error {
	return p.dao.SaveConfig(&model.Config{
		Key:   key,
		Value: value,
	})
}

func (p *PersistentConfig) IsGithub() bool {
	return p.Get(store.GithubAppID) != ""
}

func (p *PersistentConfig) IsGitlab() bool {
	return p.Get(store.GitlabClientID) != ""
}

func (p *PersistentConfig) Org() string {
	if p.IsGithub() {
		return p.Get(store.GithubOrg)
	} else if p.IsGitlab() {
		return p.Get(store.GitlabOrg)
	}

	return ""
}

func (p *PersistentConfig) ScmURL() string {
	if p.IsGithub() {
		return "https://github.com"
	} else if p.IsGitlab() {
		if p.Get(store.GitlabURL) != "" {
			return p.Get(store.GitlabURL)
		}
		return "https://gitlab.com"
	}

	return ""
}

func (p *PersistentConfig) saveConfigFile(config *Config) error {
	configs := map[string]string{}
	t := reflect.TypeOf(*config)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("envconfig")
		value := reflect.ValueOf(*config).Field(i)

		if value.Kind() == reflect.Struct {
			for j := 0; j < value.NumField(); j++ {
				nestedField := value.Type().Field(j)
				nestedTag := nestedField.Tag.Get("envconfig")
				nestedValue := value.Field(j)

				if !nestedValue.IsZero() {
					if nestedValue.Kind() == reflect.Bool {
						configs[nestedTag] = strconv.FormatBool(nestedValue.Bool())
					} else if nestedValue.Kind() == reflect.Int {
						configs[nestedTag] = strconv.FormatInt(nestedValue.Int(), 10)
					} else {
						configs[nestedTag] = nestedValue.String()
					}
				}
			}
		} else {
			if !value.IsZero() {
				if value.Kind() == reflect.Bool {
					configs[tag] = strconv.FormatBool(value.Bool())
				} else if value.Kind() == reflect.Int {
					configs[tag] = strconv.FormatInt(value.Int(), 10)
				} else {
					configs[tag] = value.String()
				}
			}
		}
	}

	for key, value := range configs {
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
