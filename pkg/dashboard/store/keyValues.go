package store

import (
	database_sql "database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

// SaveKeyValue sets a setting
func (db *Store) SaveKeyValue(setting *model.KeyValue) error {
	storedSetting, err := db.KeyValue(setting.Key)

	if err != nil {
		switch err {
		case database_sql.ErrNoRows:
			return meddler.Insert(db, "key_values", setting)
		default:
			return err
		}
	}

	storedSetting.Value = setting.Value
	return meddler.Update(db, "key_values", storedSetting)
}

// KeyValue returns the value of a given KeyValue key
func (db *Store) KeyValue(key string) (*model.KeyValue, error) {
	stmt := sql.Stmt(db.driver, sql.SelectKeyValue)
	data := new(model.KeyValue)
	err := meddler.QueryRow(db, data, stmt, key)
	return data, err
}
