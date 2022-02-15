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

package sql

const Dummy = "dummy"
const SelectUserByLogin = "select-user-by-login"
const SelectCommitsByRepo = "select-commits-by-repo"
const SelectKeyValue = "select-key-value"

var queries = map[string]map[string]string{
	"sqlite3": {
		Dummy: `
SELECT 1;
`,
		SelectUserByLogin: `
SELECT id, login, name, email, access_token, refresh_token, expires, secret, repos, favorite_repos, favorite_services
FROM users
WHERE login = ?;
`,
		SelectCommitsByRepo: `
SELECT id, repo, sha, url, author, author_pic, message, created, tags, status
FROM commits
WHERE repo = ?
LIMIT 20;
`,
		SelectKeyValue: `
SELECT id, key, value
FROM key_values
WHERE key = ?;
`,
	},
	"postgres": {
		Dummy: `
SELECT 1;
`,
		SelectUserByLogin: `
SELECT id, login, name, email, access_token, refresh_token, expires, secret, repos, favorite_repos, favorite_services
FROM users
WHERE login = $1;
`,
		SelectCommitsByRepo: `
SELECT id, repo, sha, url, author, author_pic, message, created, tags, status
FROM commits
WHERE repo = $1
LIMIT 20;
`,
		SelectKeyValue: `
SELECT id, key, value
FROM key_values
WHERE key = $1;
`,
	},
}
