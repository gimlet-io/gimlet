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
const createTableEvents = "create-table-events"
const addGitopsStatusColumnToEventsTable = "add-gitops_status-to-events-table"
const addResultsColumnToEventsTable = "add-results-to-events-table"
const createTableGitopsCommits = "create-table-gitopsCommits"
const addCreatedColumnToGitopsCommitsTable = "add-created-to-gitops-commits-table"
const addEnvColumnToGitopsCommitsTable = "add-env-to-gitops-commits-table"
const createTableKeyValues = "create-table-key-values"
const createTablePods = "create-table-pods"
const defaultValueForCreated = "default-value-for-created"
const defaultValueForEnv = "default-value-for-env"

type migration struct {
	name string
	stmt string
}

var migrations = map[string][]migration{
	"sqlite3": {
		{
			name: createTableUsers,
			stmt: `
CREATE TABLE IF NOT EXISTS users (
id           INTEGER PRIMARY KEY AUTOINCREMENT,
login         TEXT,
secret        TEXT,
admin         BOOLEAN,
UNIQUE(login)
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
UNIQUE(id)
);
`,
		},
		{
			name: addGitopsStatusColumnToEventsTable,
			stmt: `ALTER TABLE events ADD COLUMN gitops_hashes TEXT DEFAULT '[]';`,
		},
		{
			name: addResultsColumnToEventsTable,
			stmt: `ALTER TABLE events ADD COLUMN results TEXT DEFAULT '[]';`,
		},
		{
			name: createTableGitopsCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS gitops_commits (
id          INTEGER PRIMARY KEY AUTOINCREMENT,
sha         TEXT,
status      TEXT,
status_desc TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: addCreatedColumnToGitopsCommitsTable,
			stmt: `ALTER TABLE gitops_commits ADD COLUMN created INTEGER;`,
		},
		{
			name: addEnvColumnToGitopsCommitsTable,
			stmt: `ALTER TABLE gitops_commits ADD COLUMN env TEXT;`,
		},
		{
			name: defaultValueForCreated,
			stmt: `update gitops_commits set created=0 where created is null;`,
		},
		{
			name: defaultValueForEnv,
			stmt: `update gitops_commits set env='' where env is null;`,
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
			name: createTablePods,
			stmt: `
CREATE TABLE IF NOT EXISTS pods (
id          		  INTEGER PRIMARY KEY AUTOINCREMENT,
name				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
alert_state 		  TEXT,
alert_state_timestamp INTEGER,
UNIQUE(id)
);
`,
		},
	},
	"postgres": {
		{
			name: createTableUsers,
			stmt: `
CREATE TABLE IF NOT EXISTS users (
id           SERIAL,
login         TEXT,
secret        TEXT,
admin         BOOLEAN,
UNIQUE(login)
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
UNIQUE(id)
);
`,
		},
		{
			name: addGitopsStatusColumnToEventsTable,
			stmt: `ALTER TABLE events ADD COLUMN gitops_hashes TEXT DEFAULT '[]';`,
		},
		{
			name: addResultsColumnToEventsTable,
			stmt: `ALTER TABLE events ADD COLUMN results TEXT DEFAULT '[]';`,
		},
		{
			name: createTableGitopsCommits,
			stmt: `
CREATE TABLE IF NOT EXISTS gitops_commits (
id          SERIAL,
sha         TEXT,
status      TEXT,
status_desc TEXT,
UNIQUE(id)
);
`,
		},
		{
			name: addCreatedColumnToGitopsCommitsTable,
			stmt: `ALTER TABLE gitops_commits ADD COLUMN created INTEGER;`,
		},
		{
			name: addEnvColumnToGitopsCommitsTable,
			stmt: `ALTER TABLE gitops_commits ADD COLUMN env TEXT;`,
		},
		{
			name: defaultValueForCreated,
			stmt: `update gitops_commits set created=0 where created is null;`,
		},
		{
			name: defaultValueForEnv,
			stmt: `update gitops_commits set env='' where env is null;`,
		},
		{
			name: createTableKeyValues,
			stmt: `
CREATE TABLE IF NOT EXISTS key_values (
	id        SERIAL,
	key       TEXT,
	value      TEXT,
	UNIQUE(key)
	);
`,
		},
		{
			name: createTablePods,
			stmt: `
CREATE TABLE IF NOT EXISTS pods (
id          		  SERIAL,
name  				  TEXT,
status      		  TEXT,
status_desc 		  TEXT,
alert_state 		  TEXT,
alert_state_timestamp INTEGER,
UNIQUE(id)
);
`,
		},
	},
	"mysql": {},
}
