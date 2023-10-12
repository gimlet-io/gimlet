// Copyright 2019 Laszlo Fogas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ddl

const createTableUsers = "create-table-users"
const addNameColumnToUsersTable = "add-name-column-to-users-table"
const addAdminColumnToUsersTable = "add-admin-column-to-users-table"
const createTableEnvironments = "create-table-environments"
const createTableCommits = "create-table-commits"
const addMessageColumnToCommitsTable = "add-message-column-to-commits-table"
const addCreatedColumnToCommitsTable = "add-created-column-to-commits-table"
const defaultValueForMessage = "default-value-for-message"
const defaultValueForCreated = "default-value-for-created"
const addReposColumnToUsersTable = "addReposColumnToUsersTable"
const addFavoriteReposColumnToUsersTable = "addFavoriteReposColumnToUsersTable"
const addFavoriteServicesColumnToUsersTable = "addFavoriteServicesColumnToUsersTable"
const defaultValueForRepos = "defaultValueForRepos"
const defaultValueForFavoriteRepos = "defaultValueForFavoriteRepos"
const defaultValueForFavoriteServices = "defaultValueForFavoriteServices"
const createTableKeyValues = "create-table-key-values"
const addInfrarepoColumnToEnvironmentsTable = "addInfrarepoColumnToEnvironmentsTable"
const addAppsrepoColumnToEnvironmentsTable = "addAppsrepoColumnToEnvironmentsTable"
const defaultValueForGitopsRepos = "defaultValueForGitopsRepos"
const addRepoPerEnvColumnToEnvironmentsTable = "addRepoPerEnvColumnToEnvironmentsTable"
const defaultValueForRepoPerEnv = "defaultValueForRepoPerEnv"
const addKustomizationPerAppToEnvironmentsTable = "add-kustomization-per-app-to-environments-table"
const defaultValueForKustomizationPerApp = "default-value-for-kustomization-per-app"
const addBuiltInToEnvironmentsTable = "add-built-in-to-environments-table"
const defaultValueForBuiltInToEnvironmentsTable = "default-value-for-built-in-to-environments-table"
const createTablePods = "create-table-pods"
const createTableEvents = "create-table-events"
const createTableGitopsCommits = "create-table-gitopsCommits"
const createTableKubeEvents = "create-table-kube-events"
const createTableAlerts = "create-table-alerts"
const addPendingAtToAlertsTable = "add-pending-at-to-alerts-table"
const addFiredAtToAlertsTable = "add-fired-at-to-alerts-table"
const addResolvedAtToAlertsTable = "add-resolved-at-to-alerts-table"
const defaultTimestampsInAlertsTable = "defaultTimestampsInAlertsTable"

type migration struct {
	name string
	stmt string
}

var migrations = map[string][]migration{
	"sqlite": {
		{
			name: createTableUsers,
			stmt: `
CREATE TABLE IF NOT EXISTS users (
id           INTEGER PRIMARY KEY AUTOINCREMENT,
login         TEXT,
email         TEXT,
access_token  TEXT,
refresh_token TEXT,
expires       INT,
secret        TEXT,
UNIQUE(login)
);
`,
		},
		{
			name: addNameColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN name TEXT default '';`,
		},
		{
			name: addAdminColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN admin BOOLEAN default false;`,
		},
		{
			name: createTableEnvironments,
			stmt: `
CREATE TABLE IF NOT EXISTS environments (
id         	INTEGER PRIMARY KEY AUTOINCREMENT,
name        TEXT,
UNIQUE(name)
);
`,
		},
		{
			name: createTableCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS commits (
id         INTEGER PRIMARY KEY AUTOINCREMENT,
sha        TEXT,
url        TEXT,
author     TEXT,
author_pic TEXT,
tags       TEXT,
repo       TEXT,
status 	   TEXT,
UNIQUE(sha,repo)
);
`,
		},
		{
			name: addMessageColumnToCommitsTable,
			stmt: `ALTER TABLE commits ADD COLUMN message TEXT;`,
		},
		{
			name: addCreatedColumnToCommitsTable,
			stmt: `ALTER TABLE commits ADD COLUMN created INTEGER;`,
		},
		{
			name: defaultValueForMessage,
			stmt: `update commits set message='' where message is null;`,
		},
		{
			name: defaultValueForCreated,
			stmt: `update commits set created=0 where created is null;`,
		},
		{
			name: addReposColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN repos TEXT;`,
		},
		{
			name: addFavoriteReposColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN favorite_repos TEXT;`,
		},
		{
			name: addFavoriteServicesColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN favorite_services TEXT;`,
		},
		{
			name: defaultValueForRepos,
			stmt: `update users set repos='[]' where repos is null;`,
		},
		{
			name: defaultValueForFavoriteRepos,
			stmt: `update users set favorite_repos='[]' where favorite_repos is null;`,
		},
		{
			name: defaultValueForFavoriteServices,
			stmt: `update users set favorite_services='[]' where favorite_services is null;`,
		},
		{
			name: createTableKeyValues,
			stmt: `
CREATE TABLE IF NOT EXISTS key_values (
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	key       TEXT,
	value      TEXT,
	UNIQUE(key)
	);
`,
		},
		{
			name: addInfrarepoColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN infra_repo TEXT;`,
		},
		{
			name: addAppsrepoColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN apps_repo TEXT;`,
		},
		{
			name: defaultValueForGitopsRepos,
			stmt: `update environments set infra_repo='', apps_repo='' where infra_repo is null and apps_repo is null;`,
		},
		{
			name: addRepoPerEnvColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN repo_per_env TEXT;`,
		},
		{
			name: defaultValueForRepoPerEnv,
			stmt: `update environments set repo_per_env=false where repo_per_env is null;`,
		},
		{
			name: addKustomizationPerAppToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN kustomization_per_app BOOLEAN;`,
		},
		{
			name: defaultValueForKustomizationPerApp,
			stmt: `update environments set kustomization_per_app=true where kustomization_per_app is null;`,
		},
		{
			name: addBuiltInToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN built_in BOOLEAN;`,
		},
		{
			name: defaultValueForBuiltInToEnvironmentsTable,
			stmt: `update environments set built_in=false where built_in is null;`,
		},
		{
			name: createTablePods,
			stmt: `
CREATE TABLE IF NOT EXISTS pods (
id          		  INTEGER PRIMARY KEY AUTOINCREMENT,
name				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: createTableEvents,
			stmt: `
CREATE TABLE IF NOT EXISTS events (
id            TEXT,
created       INTEGER,
type          TEXT,
blob          TEXT,
status        TEXT DEFAULT 'new',
status_desc   TEXT DEFAULT '',
repository    TEXT,
branch        TEXT,
event         TEXT,
source_branch TEXT,
target_branch TEXT,
tag           TEXT,
sha           TEXT,
artifact_id   TEXT,
gitops_hashes TEXT DEFAULT '[]',
results		  TEXT DEFAULT '[]',
UNIQUE(id)
);
`,
		},
		{
			name: createTableGitopsCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS gitops_commits (
id          INTEGER PRIMARY KEY AUTOINCREMENT,
sha         TEXT,
status      TEXT,
status_desc TEXT,
created 	INTEGER DEFAULT 0,
env 		TEXT DEFAULT '',
UNIQUE(id)
);
`,
		},
		{
			name: createTableKubeEvents,
			stmt: `
CREATE TABLE IF NOT EXISTS kube_events (
id          		  INTEGER PRIMARY KEY AUTOINCREMENT,
name				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: createTableAlerts,
			stmt: `
CREATE TABLE IF NOT EXISTS alerts (
id				  INTEGER PRIMARY KEY AUTOINCREMENT,
type			  TEXT,
name			  TEXT,
deployment_name   TEXT,
status			  TEXT,
status_desc 	  TEXT,
last_state_change INTEGER,
count			  INTEGER,
UNIQUE(id)
);
`,
		},
		{
			name: addPendingAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN pending_at INTEGER;`,
		},
		{
			name: addFiredAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN fired_at INTEGER;`,
		},
		{
			name: addResolvedAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN resolved_at INTEGER;`,
		},
		{
			name: defaultTimestampsInAlertsTable,
			stmt: `update alerts set pending_at=0, fired_at=0, resolved_at=0 where pending_at is null;`,
		},
	},
	"postgres": {
		{
			name: createTableUsers,
			stmt: `
CREATE TABLE IF NOT EXISTS users (
id            SERIAL,
login         TEXT,
email         TEXT,
access_token  TEXT,
refresh_token TEXT,
expires       INT,
secret        TEXT,
UNIQUE(login)
);
`,
		},
		{
			name: addNameColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN name TEXT default '';`,
		},
		{
			name: addAdminColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN admin BOOLEAN;`,
		},
		{
			name: createTableEnvironments,
			stmt: `
CREATE TABLE IF NOT EXISTS environments (
id         	SERIAL,
name        TEXT,
UNIQUE(name)
);
`,
		},
		{
			name: createTableCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS commits (
id         SERIAL,
sha        TEXT,
url        TEXT,
author     TEXT,
author_pic TEXT,
tags       TEXT,
repo       TEXT,
status 	   TEXT,
UNIQUE(sha,repo)
);
`,
		},
		{
			name: addMessageColumnToCommitsTable,
			stmt: `ALTER TABLE commits ADD COLUMN message TEXT;`,
		},
		{
			name: addCreatedColumnToCommitsTable,
			stmt: `ALTER TABLE commits ADD COLUMN created INTEGER;`,
		},
		{
			name: defaultValueForMessage,
			stmt: `update commits set message='' where message is null;`,
		},
		{
			name: defaultValueForCreated,
			stmt: `update commits set created=0 where created is null;`,
		},
		{
			name: addReposColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN repos TEXT;`,
		},
		{
			name: addFavoriteReposColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN favorite_repos TEXT;`,
		},
		{
			name: addFavoriteServicesColumnToUsersTable,
			stmt: `ALTER TABLE users ADD COLUMN favorite_services TEXT;`,
		},
		{
			name: defaultValueForRepos,
			stmt: `update users set repos='[]' where repos is null;`,
		},
		{
			name: defaultValueForFavoriteRepos,
			stmt: `update users set favorite_repos='[]' where favorite_repos is null;`,
		},
		{
			name: defaultValueForFavoriteServices,
			stmt: `update users set favorite_services='[]' where favorite_services is null;`,
		},
		{
			name: createTableKeyValues,
			stmt: `
CREATE TABLE IF NOT EXISTS key_values (
	id        SERIAL,
	key       TEXT,
	value     TEXT,
	UNIQUE(key)
	);
`},
		{
			name: addInfrarepoColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN infra_repo TEXT;`,
		},
		{
			name: addAppsrepoColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN apps_repo TEXT;`,
		},
		{
			name: defaultValueForGitopsRepos,
			stmt: `update environments set infra_repo='', apps_repo='' where infra_repo is null and apps_repo is null;`,
		},
		{
			name: addRepoPerEnvColumnToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN repo_per_env TEXT;`,
		},
		{
			name: defaultValueForRepoPerEnv,
			stmt: `update environments set repo_per_env=false where repo_per_env is null;`,
		},
		{
			name: addKustomizationPerAppToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN kustomization_per_app BOOLEAN;`,
		},
		{
			name: defaultValueForKustomizationPerApp,
			stmt: `update environments set kustomization_per_app=true where kustomization_per_app is null;`,
		},
		{
			name: addBuiltInToEnvironmentsTable,
			stmt: `ALTER TABLE environments ADD COLUMN built_in BOOLEAN;`,
		},
		{
			name: defaultValueForBuiltInToEnvironmentsTable,
			stmt: `update environments set built_in=false where built_in is null;`,
		},
		{
			name: createTablePods,
			stmt: `
CREATE TABLE IF NOT EXISTS pods (
id          		  SERIAL,
name  				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: createTableEvents,
			stmt: `
CREATE TABLE IF NOT EXISTS events (
id            TEXT,
created       INTEGER,
type          TEXT,
blob          TEXT,
status        TEXT DEFAULT 'new',
status_desc   TEXT DEFAULT '',
repository    TEXT,
branch        TEXT,
event         TEXT,
source_branch TEXT,
target_branch TEXT,
tag           TEXT,
sha           TEXT,
artifact_id   TEXT,
gitops_hashes TEXT DEFAULT '[]',
results		  TEXT DEFAULT '[]',
UNIQUE(id)
);
`,
		},
		{
			name: createTableGitopsCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS gitops_commits (
id          SERIAL,
sha         TEXT,
status      TEXT,
status_desc TEXT,
created 	INTEGER DEFAULT 0,
env 		TEXT DEFAULT '',
UNIQUE(id)
);
`,
		},
		{
			name: createTableKubeEvents,
			stmt: `
CREATE TABLE IF NOT EXISTS kube_events (
id          		  SERIAL,
name				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: createTableAlerts,
			stmt: `
CREATE TABLE IF NOT EXISTS alerts (
id				  SERIAL,
type			  TEXT,
name			  TEXT,
deployment_name   TEXT,
status			  TEXT,
status_desc 	  TEXT,
last_state_change INTEGER,
count			  INTEGER,
UNIQUE(id)
);
`,
		},
		{
			name: addPendingAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN pending_at INTEGER;`,
		},
		{
			name: addFiredAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN fired_at INTEGER;`,
		},
		{
			name: addResolvedAtToAlertsTable,
			stmt: `ALTER TABLE alerts ADD COLUMN resolved_at INTEGER;`,
		},
		{
			name: defaultTimestampsInAlertsTable,
			stmt: `update alerts set pending_at=0, fired_at=0, resolved_at=0 where pending_at is null;`,
		},
	},
}
