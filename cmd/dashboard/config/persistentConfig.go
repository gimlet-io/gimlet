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
		logrus.Warnf("cannot get config from db: %s", err)
	}
	return config.Value
}

func (p *PersistentConfig) Save(key string, value string) error {
	return p.dao.SaveConfig(&model.Config{
		Key:   key,
		Value: value,
	})
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
