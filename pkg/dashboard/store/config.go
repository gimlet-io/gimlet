package store

import (
	"fmt"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

const (
	Debug = "DEBUG"
	Trace = "TRACE"

	Host      = "HOST"
	JWTSecret = "JWT_SECRET"

	GithubAppID          = "GITHUB_APP_ID"
	GithubInstallationID = "GITHUB_INSTALLATION_ID"
	GithubPrivateKey     = "GITHUB_PRIVATE_KEY"
	GithubClientID       = "GITHUB_CLIENT_ID"
	GithubClientSecret   = "GITHUB_CLIENT_SECRET"
	GithubSkipVerify     = "GITHUB_SKIP_VERIFY"
	GithubDebug          = "GITHUB_DEBUG"
	GithubOrg            = "GITHUB_ORG"

	GitlabClientID     = "GITLAB_CLIENT_ID"
	GitlabClientSecret = "GITLAB_CLIENT_SECRET" // This is a personal access token of the Gitlab admin or a Group Token
	GitlabAdminToken   = "GITLAB_ADMIN_TOKEN"
	GitlabDebug        = "GITLAB_DEBUG"
	GitlabOrg          = "GITLAB_ORG"
	GitlabURL          = "GITLAB_URL"

	DatabaseDriver           = "DATABASE_DRIVER"
	DatabaseConfig           = "DATABASE_CONFIG"
	DatabaseEncryptionKey    = "DATABASE_ENCRYPTION_KEY"
	DatabaseEncryptionKeyNew = "DATABASE_ENCRYPTION_KEY_NEW"

	NotificationsProvider       = "NOTIFICATIONS_PROVIDER"
	NotificationsToken          = "NOTIFICATIONS_TOKEN"
	NotificationsDefaultChannel = "NOTIFICATIONS_DEFAULT_CHANNEL"
	NotificationsChannelMapping = "NOTIFICATIONS_CHANNEL_MAPPING"

	ChartName    = "CHART_NAME"
	ChartRepo    = "CHART_REPO"
	ChartVersion = "CHART_VERSION"

	RepoCachePath           = "REPO_CACHE_PATH"
	WebhookSecret           = "WEBHOOK_SECRET"
	ReleaseHistorySinceDays = "RELEASE_HISTORY_SINCE_DAYS"
	BootstrapEnv            = "BOOTSTRAP_ENV"
	UserflowToken           = "USERFLOW_TOKEN"

	PrintAdminToken = "PRINT_ADMIN_TOKEN"
	AdminToken      = "ADMIN_TOKEN"

	GitopsRepo  = "GITOPS_REPO"  // Deprecated
	GitopsRepos = "GITOPS_REPOS" // Deprecated

	GitopsRepoDeployKeyPath = "GITOPS_REPO_DEPLOY_KEY_PATH"
	GitSSHAddressFormat     = "GIT_SSH_ADDRESS_FORMAT"
	ReleaseStats            = "RELEASE_STATS"

	TermsOfServiceFeatureFlag = "FEATURE_TERMS_OF_SERVICE"
)

func (db *Store) SaveConfig(config *model.Config) error {
	configFromDb, _ := db.GetConfig(config.Key)
	if configFromDb.Key != "" {
		if unchangeable(configFromDb.Key) {
			return fmt.Errorf("config with key %s already exists in db and cannot be overwritten", configFromDb.Key)
		}
		configFromDb.Value = config.Value
		return meddler.Update(db, "config", configFromDb)
	}
	return meddler.Insert(db, "config", config)
}

func (db *Store) GetConfig(key string) (*model.Config, error) {
	stmt := sql.Stmt(db.driver, sql.SelectConfigByKey)
	data := new(model.Config)
	err := meddler.QueryRow(db, data, stmt, key)
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
