package store

import (
	"fmt"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

const (
	GithubAppID          = "GITHUB_APP_ID"
	GithubInstallationID = "GITHUB_INSTALLATION_ID"
	GithubPrivateKey     = "GITHUB_PRIVATE_KEY"
	GithubClientID       = "GITHUB_CLIENT_ID"
	GithubClientSecret   = "GITHUB_CLIENT_SECRET"
	GithubOrg            = "GITHUB_ORG"

	GitlabClientID     = "GITLAB_CLIENT_ID"
	GitlabClientSecret = "GITLAB_CLIENT_SECRET"
	GitlabAdminToken   = "GITLAB_ADMIN_TOKEN"
	GitlabOrg          = "GITLAB_ORG"
	GitlabURL          = "GITLAB_URL"
)

func (db *Store) SaveConfig(config *model.Config) error {
	configFromDb, _ := db.getConfig(config.Key)
	if configFromDb.Key != "" {
		if unchangeable(configFromDb.Key) {
			return fmt.Errorf("config with key %s already exists in db and cannot be overwritten", configFromDb.Key)
		}
		configFromDb.Value = config.Value
		return meddler.Update(db, "config", configFromDb)
	}
	return meddler.Insert(db, "config", config)
}

func (db *Store) getConfig(key string) (*model.Config, error) {
	stmt := sql.Stmt(db.driver, sql.SelectConfigByKey)
	data := new(model.Config)
	err := meddler.QueryRow(db, data, stmt, key)
	return data, err
}

func (db *Store) GetConfigs() ([]*model.Config, error) {
	stmt := sql.Stmt(db.driver, sql.SelectConfigs)
	var data []*model.Config
	err := meddler.QueryAll(db, &data, stmt)
	return data, err
}

func unchangeable(key string) bool {
	return key == GithubAppID ||
		key == GithubInstallationID ||
		key == GithubPrivateKey ||
		key == GithubClientID ||
		key == GithubClientSecret ||
		key == GithubOrg ||
		key == GitlabClientID ||
		key == GitlabClientSecret ||
		key == GitlabAdminToken ||
		key == GitlabOrg ||
		key == GitlabURL
}
