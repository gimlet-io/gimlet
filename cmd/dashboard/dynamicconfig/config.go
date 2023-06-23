package dynamicconfig

import (
	"database/sql"
	"encoding/json"
	"reflect"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/kelseyhightower/envconfig"
)

const CONFIG_STORAGE_KEY = "config"

// LoadDynamicConfig persist env config in the db and loads config from the db.
// DB values take precedence
func LoadDynamicConfig(dao *store.Store) (*DynamicConfig, error) {
	// First we load config from the environment
	cfg := &DynamicConfig{
		dao: dao,
	}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	// Then we persist it to the database if they are not set yet
	err = cfg.persistEnvConfig()
	if err != nil {
		return nil, err
	}

	// Then we reload from the DB, so DB values take precedence
	err = cfg.load()
	return cfg, err
}

// Config holds Gimlet configuration that is stored in the database
// * It can be initiatied with environment variables
// * and dynamically changed runtime that is persisted in the database.
// * Values in the database take precedence.
//
// We have a single instance of this struct in Gimlet
// changes to this struct are reflected application wide as it has a pointer reference.
// To make the changes persistent, call Persist()
type DynamicConfig struct {
	dao *store.Store

	DummyString  string
	DummyString2 string
	DummyBool    bool
	DummyBool2   bool
	DummyInt     int

	Github config.Github
	Gitlab config.Gitlab

	JWTSecret string
	AdminKey  string
}

// persist all config fields that are not already set in the database
func (c *DynamicConfig) persistEnvConfig() error {
	var persistedConfig DynamicConfig
	persistedConfigString, err := c.dao.KeyValue(CONFIG_STORAGE_KEY)
	if err == nil {
		err = json.Unmarshal([]byte(persistedConfigString.Value), &persistedConfig)
		if err != nil {
			return err
		}
	} else if err == sql.ErrNoRows {
		persistedConfig = DynamicConfig{}
	} else if err != nil {
		return err
	}

	updateConfigWhenZeroValue(&persistedConfig, c)
	persistedConfig.dao = c.dao

	return persistedConfig.Persist()
}

func updateConfigWhenZeroValue(toUpdate *DynamicConfig, new *DynamicConfig) {
	t := reflect.TypeOf(*toUpdate)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := reflect.ValueOf(*toUpdate).Field(i)
		newValue := reflect.ValueOf(*new).FieldByName(field.Name)

		if value.Kind() == reflect.Struct {
			for j := 0; j < value.NumField(); j++ {
				nestedField := value.Type().Field(j)
				nestedValue := value.Field(j)
				newNestedValue := newValue.FieldByName(nestedField.Name)

				if nestedValue.IsZero() {
					obj := reflect.ValueOf(toUpdate).Elem()
					if nestedValue.Kind() == reflect.Bool {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetBool(newNestedValue.Bool())
					} else if nestedValue.Kind() == reflect.Int {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetInt(newNestedValue.Int())
					} else if nestedValue.Kind() == reflect.String {
						obj.FieldByName(field.Name).FieldByName(nestedField.Name).SetString(newNestedValue.String())
					}
				}
			}
		} else {
			obj := reflect.ValueOf(toUpdate).Elem()

			if value.IsZero() {
				if value.Kind() == reflect.Bool {
					obj.FieldByName(field.Name).SetBool(newValue.Bool())
				} else if value.Kind() == reflect.Int {
					obj.FieldByName(field.Name).SetInt(newValue.Int())
				} else if value.Kind() == reflect.String {
					obj.FieldByName(field.Name).SetString(newValue.String())
				}
			}
		}
	}
}

// Persist saves the config struct to the DB
// When we are updating configs, like regsistering new github applications,
// the config user should call a Persist on the config
func (c *DynamicConfig) Persist() error {
	configString, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return c.dao.SaveKeyValue(&model.KeyValue{
		Key:   CONFIG_STORAGE_KEY,
		Value: string(configString),
	})
}

func (c *DynamicConfig) load() error {
	configString, err := c.dao.KeyValue(CONFIG_STORAGE_KEY)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(configString.Value), c)
}

func (c *DynamicConfig) IsGithub() bool {
	return c.Github.AppID != ""
}

func (c *DynamicConfig) IsGitlab() bool {
	return c.Gitlab.ClientID != ""
}

func (c *DynamicConfig) Org() string {
	if c.IsGithub() {
		return c.Github.Org
	} else if c.IsGitlab() {
		return c.Gitlab.Org
	}

	return ""
}

func (c *DynamicConfig) ScmURL() string {
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
