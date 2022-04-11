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
const SelectAllUser = "select-all-user"
const DeleteUser = "deleteUser"
const SelectUnprocessedEvents = "select-unprocessed-events"
const UpdateEventStatus = "update-event-status"
const SelectGitopsCommitBySha = "select-gitops-commit-by-sha"
const SelectGitopsCommits = "select-gitops-commits"
const SelectKeyValue = "select-key-value"

var queries = map[string]map[string]string{
	"sqlite3": {
		Dummy: `
SELECT 1;
`,
		SelectUserByLogin: `
SELECT id, login, secret, admin
FROM users
WHERE login = $1;
`,
		SelectAllUser: `
SELECT id, login, secret, admin
FROM users;
`,
		DeleteUser: `
DELETE FROM users where login = $1;
`,
		SelectUnprocessedEvents: `
SELECT id, created, type, blob, status, status_desc, sha, repository, branch, event, source_branch, target_branch, tag, artifact_id
FROM events
WHERE status='new' order by created ASC limit 10;
`,
		UpdateEventStatus: `
UPDATE events SET status = $1, status_desc = $2, gitops_hashes = $3 WHERE id = $4;
`,
		SelectGitopsCommitBySha: `
SELECT id, sha, status, status_desc
FROM gitops_commits
WHERE sha = $1;
`,
		SelectGitopsCommits: `
SELECT id, sha, status, status_desc
FROM gitops_commits
LIMIT 20;
`,
		SelectKeyValue: `
SELECT id, key, value
FROM key_values
WHERE key = $1;
`,
	},
	"postgres": {
		Dummy: `
SELECT 1;
`,
		SelectUserByLogin: `
SELECT id, login, secret, admin
FROM users
WHERE login = $1;
`,
		SelectAllUser: `
SELECT id, login, secret, admin
FROM users;
`,
		DeleteUser: `
DELETE FROM users where login = $1;
`,
		SelectUnprocessedEvents: `
SELECT id, created, type, blob, status, status_desc, sha, repository, branch, event, source_branch, target_branch, tag, artifact_id
FROM events
WHERE status='new' order by created ASC limit 10;
`,
		UpdateEventStatus: `
UPDATE events SET status = $1, status_desc = $2, gitops_hashes = $3 WHERE id = $4;
`,
		SelectGitopsCommitBySha: `
SELECT id, sha, status, status_desc
FROM gitops_commits
WHERE sha = $1;
`,
		SelectGitopsCommits: `
SELECT id, sha, status, status_desc
FROM gitops_commits
LIMIT 20;
`,
		SelectKeyValue: `
SELECT id, key, value
FROM key_values
WHERE key = $1;
`,
	},
	"mysql": {},
}
