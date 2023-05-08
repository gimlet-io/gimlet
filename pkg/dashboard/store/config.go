package store

import (
	"fmt"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) SaveConfig(config *model.Config) error {
	valueFromDb, _ := db.GetConfigValue(config.Key)
	if valueFromDb != "" {
		return fmt.Errorf("config with key %s already exists in db and cannot be overwritten", config.Key)
	}
	return meddler.Insert(db, "config", config)
}

func (db *Store) GetConfigValue(key string) (string, error) {
	stmt := sql.Stmt(db.driver, sql.SelectConfigByKey)
	data := new(model.Config)
	err := meddler.QueryRow(db, data, stmt, key)
	return data.Value, err
}
