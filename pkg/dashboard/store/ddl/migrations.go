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
`,
		},
	},
}
