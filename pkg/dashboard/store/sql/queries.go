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
const SelectEnvironments = "select-environments"
const SelectEnvironment = "select-environment"
const DeleteEnvironment = "delete-environment"
const SelectPendingPods = "select-pending-pods"
const SelectPodByName = "select-pod-by-name"
const DeletePodByName = "delete-pod-by-name"
const SelectUnprocessedEvents = "select-unprocessed-events"
const UpdateEventStatus = "update-event-status"
const SelectGitopsCommitBySha = "select-gitops-commit-by-sha"
const SelectGitopsCommits = "select-gitops-commits"

var queries = map[string]map[string]string{
	"sqlite3": {
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
		SelectEnvironments: `
SELECT id, name, infra_repo, apps_repo, repo_per_env
FROM environments
ORDER BY name asc;
`,
		SelectEnvironment: `
SELECT id, name, infra_repo, apps_repo, repo_per_env
FROM environments
WHERE name = $1;
`,
		DeleteEnvironment: `
DELETE FROM environments
WHERE name = ?;
`,
		SelectPendingPods: `
SELECT id, name, status, status_desc, alert_state, alert_state_timestamp
FROM pods
WHERE alert_state LIKE 'Pending';
`,
		SelectPodByName: `
SELECT id, name, status, status_desc, alert_state, alert_state_timestamp
FROM pods
WHERE name = $1;
`,
		DeletePodByName: `
DELETE FROM pods where name = $1;
`,
		SelectUnprocessedEvents: `
SELECT id, created, type, blob, status, status_desc, sha, repository, branch, event, source_branch, target_branch, tag, artifact_id
FROM events
WHERE status='new' order by created ASC limit 10;
`,
		UpdateEventStatus: `
UPDATE events SET status = $1, status_desc = $2, results = $3 WHERE id = $4;
`,
		SelectGitopsCommitBySha: `
SELECT id, sha, status, status_desc, created
FROM gitops_commits
WHERE sha = $1;
`,
		SelectGitopsCommits: `
SELECT id, sha, status, status_desc, created, env
FROM gitops_commits
ORDER BY created DESC
LIMIT 20;
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
		SelectEnvironments: `
SELECT id, name, infra_repo, apps_repo, repo_per_env
FROM environments
ORDER BY name asc;
`,
		SelectEnvironment: `
SELECT id, name, infra_repo, apps_repo, repo_per_env
FROM environments
WHERE name = $1;
`,
		DeleteEnvironment: `
DELETE FROM environments
WHERE name = $1;
`,
		SelectPendingPods: `
SELECT id, name, status, status_desc, alert_state, alert_state_timestamp
FROM pods
WHERE alert_state LIKE 'Pending';
`,
		SelectPodByName: `
SELECT id, name, status, status_desc, alert_state, alert_state_timestamp
FROM pods
WHERE name = $1;
`,
		DeletePodByName: `
DELETE FROM pods where name = $1;
`,
		SelectUnprocessedEvents: `
SELECT id, created, type, blob, status, status_desc, sha, repository, branch, event, source_branch, target_branch, tag, artifact_id
FROM events
WHERE status='new' order by created ASC limit 10;
`,
		UpdateEventStatus: `
UPDATE events SET status = $1, status_desc = $2, results = $3 WHERE id = $4;
`,
		SelectGitopsCommitBySha: `
SELECT id, sha, status, status_desc, created
FROM gitops_commits
WHERE sha = $1;
`,
		SelectGitopsCommits: `
SELECT id, sha, status, status_desc, created, env
FROM gitops_commits
ORDER BY created DESC
LIMIT 20;
`,
	},
}
