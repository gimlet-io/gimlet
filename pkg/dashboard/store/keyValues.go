package store

import (
	database_sql "database/sql"
	"encoding/json"
	"fmt"
	"time"

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

func (db *Store) ReposWithCleanupPolicy() ([]string, error) {
	reposWithCleanupPolicyKeyValue, err := db.KeyValue(model.ReposWithCleanupPolicy)
	if err != nil {
		return []string{}, err
	}
	var reposWithCleanupPolicy []string
	err = json.Unmarshal([]byte(reposWithCleanupPolicyKeyValue.Value), &reposWithCleanupPolicy)
	if err != nil {
		return []string{}, err
	}

	return reposWithCleanupPolicy, nil
}

func (db *Store) SaveReposWithCleanupPolicy(reposWithCleanupPolicy []string) error {
	reposWithCleanupPolicyBytes, err := json.Marshal(reposWithCleanupPolicy)
	if err != nil {
		return err
	}

	reposWithCleanupPolicyKeyValue := &model.KeyValue{
		Key:   model.ReposWithCleanupPolicy,
		Value: string(reposWithCleanupPolicyBytes),
	}

	return db.SaveKeyValue(reposWithCleanupPolicyKeyValue)
}

func (db *Store) DeploymentSilencedUntil(deployment string, alertType string) (int64, error) {
	object := fmt.Sprintf("%s-%s", deployment, alertType)
	silencedUntil, err := db.KeyValue(object)
	if err != nil {
		return 0, err
	}

	var until *time.Time
	t, err := time.Parse(time.RFC3339, silencedUntil.Value)
	if err != nil {
		return 0, fmt.Errorf("cannot parse until date: %s", err)
	}
	until = &t

	return until.Unix(), nil
}
