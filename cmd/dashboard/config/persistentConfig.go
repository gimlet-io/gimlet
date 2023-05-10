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
	err := p.Save(config)
	return p, err
}

func (p *PersistentConfig) Get() (*Config, error) {
	configsFromDb, err := p.dao.GetConfigs()
	if err != nil {
		logrus.Error("cannot get configs from db")
	}

	config := &Config{}
	t := reflect.TypeOf(*config)
	for _, v := range configsFromDb {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("envconfig")
			value := reflect.ValueOf(config).Elem().Field(i)

			if value.Kind() == reflect.Struct {
				for j := 0; j < value.NumField(); j++ {
					nestedField := value.Type().Field(j)
					nestedTag := nestedField.Tag.Get("envconfig")
					nestedValue := value.FieldByIndex([]int{j})

					if nestedTag == v.Key {
						if nestedValue.Kind() == reflect.Bool {
							boolValue, _ := strconv.ParseBool(v.Value)
							nestedValue.SetBool(boolValue)
						} else if nestedValue.Kind() == reflect.Int {
							intValue, _ := strconv.ParseInt(v.Value, 10, 64)
							nestedValue.SetInt(intValue)
						} else {
							nestedValue.SetString(v.Value)
						}
					}
				}
			} else {
				if tag == v.Key {
					if value.Kind() == reflect.Bool {
						boolValue, _ := strconv.ParseBool(v.Value)
						value.SetBool(boolValue)
					} else if value.Kind() == reflect.Int {
						intValue, _ := strconv.ParseInt(v.Value, 10, 64)
						value.SetInt(intValue)
					} else {
						value.SetString(v.Value)
					}
				}
			}
		}
	}

	return config, nil
}

func (p *PersistentConfig) Save(config *Config) error {
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
