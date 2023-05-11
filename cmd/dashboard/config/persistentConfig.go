package config

import (
	"fmt"
	"reflect"

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
		logrus.Error("cannot get config from db: %s", err)
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
	configs := make(map[string]string)

	v := reflect.ValueOf(config).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("envconfig")

		if !field.IsZero() {
			switch field.Kind() {
			case reflect.Bool, reflect.Int:
				configs[tag] = fmt.Sprintf("%v", field.Interface())
			default:
				configs[tag] = field.String()
			}
		}

		if field.Kind() == reflect.Struct {
			for j := 0; j < field.NumField(); j++ {
				nestedField := field.Field(j)
				nestedTag := t.Field(i).Tag.Get("envconfig")

				if !nestedField.IsZero() {
					switch nestedField.Kind() {
					case reflect.Bool, reflect.Int:
						configs[nestedTag] = fmt.Sprintf("%v", nestedField.Interface())
					default:
						configs[nestedTag] = nestedField.String()
					}
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
