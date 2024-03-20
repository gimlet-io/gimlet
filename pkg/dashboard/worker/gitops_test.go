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

package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_gitopsTemplateAndWrite(t *testing.T) {
	var a dx.Artifact
	json.Unmarshal([]byte(`
{
  "version": {
    "repositoryName": "my-app",
    "sha": "ea9ab7cc31b2599bf4afcfd639da516ca27a4780",
    "branch": "master",
    "authorName": "Jane Doe",
    "authorEmail": "jane@doe.org",
    "committerName": "Jane Doe",
    "committerEmail": "jane@doe.org",
    "message": "Bugfix 123",
    "url": "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780"
  },
  "environments": [
    {
      "App": "my-app",
      "Env": "staging",
      "Namespace": "staging",
      "Deploy": {
        "Branch": "master",
        "Event": "push"
      },
      "Chart": {
        "Repository": "https://chart.onechart.dev",
        "Name": "onechart",
        "Version": "0.21.0"
      },
      "Values": {
        "image": {
          "repository": "ghcr.io/gimlet-io/my-app",
          "tag": "{{ .GITHUB_SHA }}"
        },
        "replicas": 1,
        "volumes": [
		  {
            "name": "uploads",
			"path": "/files",
			"size": "12Gi",
			"storageClass": "gp3"
          },
		  {
            "name": "errors",
			"path": "/errors",
			"size": "12Gi",
			"storageClass": "gp3"
          }
		]
      }
    }
  ],
  "items": [
    {
      "name": "CI",
      "url": "https://jenkins.example.com/job/dev/84/display/redirect"
    }
  ]
}
`), &a)

	repo, _ := git.Init(memory.NewStorage(), memfs.New())
	repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{""}})

	repoPerEnv := false
	_, err := gitopsTemplateAndWrite(repo, a.Environments[0], &dx.Release{}, "", repoPerEnv, nil, nil, nil, nil)
	assert.Nil(t, err)
	content, _ := nativeGit.Content(repo, "staging/my-app/deployment.yaml")
	assert.True(t, len(content) > 100)
	content, _ = nativeGit.Content(repo, "staging/my-app/release.json")
	assert.True(t, len(content) > 1)
	content, _ = nativeGit.Content(repo, "staging/release.json")
	assert.True(t, len(content) > 1)

	repoPerEnv = true
	_, err = gitopsTemplateAndWrite(repo, a.Environments[0], &dx.Release{}, "", repoPerEnv, nil, nil, nil, nil)
	assert.Nil(t, err)
	content, _ = nativeGit.Content(repo, "my-app/deployment.yaml")
	assert.True(t, len(content) > 100)
	content, _ = nativeGit.Content(repo, "my-app/release.json")
	assert.True(t, len(content) > 1)
	content, _ = nativeGit.Content(repo, "release.json")
	assert.True(t, len(content) > 1)
}

func Test_gitopsTemplateAndWrite_deleteStaleFiles(t *testing.T) {
	repo, _ := git.Init(memory.NewStorage(), memfs.New())
	repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{""}})
	var a dx.Artifact

	withVolume := `
{
  "version": {
    "repositoryName": "my-app",
    "sha": "ea9ab7cc31b2599bf4afcfd639da516ca27a4780",
    "branch": "master",
    "authorName": "Jane Doe",
    "authorEmail": "jane@doe.org",
    "committerName": "Jane Doe",
    "committerEmail": "jane@doe.org",
    "message": "Bugfix 123",
    "url": "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780"
  },
  "environments": [
    {
      "App": "my-app",
      "Env": "staging",
      "Namespace": "staging",
      "Chart": {
        "Repository": "https://chart.onechart.dev",
        "Name": "onechart",
        "Version": "0.21.0"
      },
      "Values": {
        "volumes": [
		  {
            "name": "uploads",
			"path": "/files",
			"size": "12Gi",
			"storageClass": "gp3"
          }
		]
      }
    }
  ]
}
`

	json.Unmarshal([]byte(withVolume), &a)

	repoPerEnv := true
	_, err := gitopsTemplateAndWrite(repo, a.Environments[0], &dx.Release{}, "", repoPerEnv, nil, nil, nil, nil)
	assert.Nil(t, err)

	_, err = gitopsTemplateAndWrite(repo, a.Environments[0], &dx.Release{}, "", repoPerEnv, nil, nil, nil, nil)
	assert.Nil(t, err)

	content, _ := nativeGit.Content(repo, "my-app/deployment.yaml")
	assert.True(t, len(content) > 100)
	content, _ = nativeGit.Content(repo, "my-app/pvc.yaml")
	assert.True(t, len(content) > 100)

	withoutVolume := `
{
  "version": {
    "repositoryName": "my-app",
    "sha": "ea9ab7cc31b2599bf4afcfd639da516ca27a4780",
    "branch": "master",
    "authorName": "Jane Doe",
    "authorEmail": "jane@doe.org",
    "committerName": "Jane Doe",
    "committerEmail": "jane@doe.org",
    "message": "Bugfix 123",
    "url": "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780"
  },
  "environments": [
    {
      "App": "my-app",
      "Env": "staging",
      "Namespace": "staging",
      "Chart": {
        "Repository": "https://chart.onechart.dev",
        "Name": "onechart",
        "Version": "0.21.0"
      }
    }
  ]
}
`

	var b dx.Artifact
	json.Unmarshal([]byte(withoutVolume), &b)
	_, err = gitopsTemplateAndWrite(repo, b.Environments[0], &dx.Release{}, "", false, nil, nil, nil, nil)
	assert.Nil(t, err)

	content, _ = nativeGit.Content(repo, "staging/my-app/pvc.yaml")
	assert.Equal(t, content, "")
}

func Test_emptyTrigger(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{}, nil)
	assert.False(t, triggered, "Empty deploy policy should not trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{}, &dx.Deploy{})
	assert.False(t, triggered, "Empty deploy policy should not trigger a deploy")
}

func Test_branchTrigger(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "master",
			},
		},
		&dx.Deploy{
			Branch: "notMaster",
			Event:  dx.PushPtr(),
		})
	assert.False(t, triggered, "Branch mismatch should not trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "master",
			},
		},
		&dx.Deploy{
			Branch: "master",
			Event:  dx.PushPtr(),
		})
	assert.True(t, triggered, "Matching branch should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "master",
				Event:  *dx.PRPtr(),
			},
		},
		&dx.Deploy{
			Branch: "master",
			Event:  dx.PRPtr(),
		})
	assert.True(t, triggered, "Matching branch should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "master",
			},
		},
		&dx.Deploy{
			Branch: "master",
		})
	assert.False(t, triggered, "Branch triggers need an event always to trigger a deploy")
}

func Test_eventTrigger(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{},
		&dx.Deploy{
			Event: dx.PushPtr(),
		})
	assert.True(t, triggered, "Default Push event should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{},
		&dx.Deploy{})
	assert.False(t, triggered, "Non matching event should not trigger a deploy, default is Push in the Artifact")

	triggered = deployTrigger(
		&dx.Artifact{},
		&dx.Deploy{
			Event: dx.PRPtr(),
		})
	assert.False(t, triggered, "Non matching event should not trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{Version: dx.Version{
			Event: dx.PR,
		}},
		&dx.Deploy{
			Event: dx.PRPtr(),
		})
	assert.True(t, triggered, "Should trigger a PR deploy")

	triggered = deployTrigger(
		&dx.Artifact{Version: dx.Version{
			Event: dx.Tag,
		}},
		&dx.Deploy{
			Event: dx.TagPtr(),
		})
	assert.True(t, triggered, "Should trigger a tag deploy")
}

func Test_tag_and_branch_pattern_triggers(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "feature/coolness",
				Event:  *dx.PRPtr(),
			},
		},
		&dx.Deploy{
			Branch: "feature/*",
			Event:  dx.PRPtr(),
		})
	assert.True(t, triggered, "Matching branch pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Tag:   "v3.0.1",
				Event: *dx.TagPtr(),
			},
		},
		&dx.Deploy{
			Tag:   "v*",
			Event: dx.TagPtr(),
		})
	assert.True(t, triggered, "Matching tag pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Tag: "xxx",
			},
		},
		&dx.Deploy{
			Tag:   "v*",
			Event: dx.TagPtr(),
		})
	assert.False(t, triggered, "Non matching tag pattern should not trigger a deploy")
}

func Test_negative_tag_and_branch_triggers(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "a-bugfix",
				Event:  *dx.PushPtr(),
			},
		},
		&dx.Deploy{
			Branch: "!main",
			Event:  dx.PushPtr(),
		})
	assert.True(t, triggered, "Matching branch pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Tag:   "v2",
				Event: *dx.TagPtr(),
			},
		},
		&dx.Deploy{
			Tag:   "!v1",
			Event: dx.TagPtr(),
		})
	assert.True(t, triggered, "Matching tag pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch: "main",
			},
		},
		&dx.Deploy{
			Branch: "!main",
			Event:  dx.TagPtr(),
		})
	assert.False(t, triggered, "Non matching branch pattern should not trigger a deploy")
}

func Test_unmarshal(t *testing.T) {
	var many dx.Manifest
	err := yaml.Unmarshal([]byte(`
app: hello
deploy:
  branch: feature/*
  event: push
cleanup:
  branch: feature/*
  event: branchDeleted
`), &many)
	assert.Nil(t, err)
	assert.NotNil(t, many.Cleanup)
	assert.NotNil(t, many.Deploy)

	a := &dx.Artifact{
		Environments: []*dx.Manifest{&many},
	}
	assert.True(t, a.HasCleanupPolicy())
}

func Test_revertTo(t *testing.T) {
	path, _ := ioutil.TempDir("", "gitops-")
	defer os.RemoveAll(path)

	repo, _ := git.PlainInit(path, false)
	initHistory(repo)

	var SHAs []string
	commits, _ := repo.Log(&git.LogOptions{})
	commits.ForEach(func(c *object.Commit) error {
		SHAs = append(SHAs, c.Hash.String())
		return nil
	})

	err := revertTo(
		"staging",
		"my-app",
		false,
		repo,
		path,
		SHAs[2],
	)
	assert.Nil(t, err)
	content, _ := nativeGit.Content(repo, "staging/my-app/file")
	assert.Equal(t, "1\n", content)

	SHAs = []string{}
	commits, _ = repo.Log(&git.LogOptions{})
	commits.ForEach(func(c *object.Commit) error {
		SHAs = append(SHAs, c.Hash.String())
		return nil
	})
	assert.Equal(t, 6, len(SHAs))

	err = revertTo(
		"staging",
		"my-app",
		false,
		repo,
		path,
		SHAs[4],
	)
	assert.Nil(t, err)
	content, _ = nativeGit.Content(repo, "staging/my-app/file")
	assert.Equal(t, "1\n", content)

	err = revertTo(
		"staging",
		"my-app",
		false,
		repo,
		path,
		SHAs[5],
	)
	assert.Nil(t, err)
	content, _ = nativeGit.Content(repo, "staging/my-app/file")
	assert.Equal(t, "0\n", content)
}

func initHistory(repo *git.Repository) {
	sha, _ := nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `0`,
			"staging/my-app/release.json": "",
			"staging/release.json":        "",
		},
		[]string{"staging/my-app"},
		"0st commit",
	)
	fmt.Printf("%s - %s\n", sha, "0")
	sha, _ = nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `1`,
			"staging/my-app/release.json": "",
			"staging/release.json":        "",
		},
		[]string{"staging/my-app"},
		"1st commit",
	)
	fmt.Printf("%s - %s\n", sha, "1")
	sha, _ = nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `2`,
			"staging/my-app/release.json": "",
			"staging/release.json":        "",
		},
		[]string{"staging/my-app"},
		"2nd commit",
	)
	fmt.Printf("%s - %s\n", sha, "2")
	sha, _ = nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `3`,
			"staging/my-app/release.json": "",
			"staging/release.json":        "",
		},
		[]string{"staging/my-app"},
		"3rd commit",
	)
	fmt.Printf("%s - %s\n", sha, "3")
}

func Test_cleanupTrigger(t *testing.T) {
	triggered := cleanupTrigger("feature/test-case-1", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Branch:       "feature/*",
		Event:        dx.BranchDeleted,
	})
	assert.True(t, triggered, "Should trigger on branch pattern")

	triggered = cleanupTrigger("fix1", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Branch:       "feature/*",
		Event:        dx.BranchDeleted,
	})
	assert.False(t, triggered, "Should not trigger on non matching branch pattern")

	triggered = cleanupTrigger("fix1", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Branch:       "preview-test",
		Event:        dx.BranchDeleted,
	})
	assert.False(t, triggered, "Should not trigger on non matching branch")

	triggered = cleanupTrigger("preview-test", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Branch:       "preview-test",
	})
	assert.True(t, triggered, "Should trigger on matching branch")

	triggered = cleanupTrigger("preview-test", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Branch:       "!main",
	})
	assert.True(t, triggered, "Should trigger on matching negated branch")

	triggered = cleanupTrigger("preview-test", &dx.Cleanup{
		AppToCleanup: "app-{{ .BRANCH }}",
		Event:        dx.BranchDeleted,
	})
	assert.False(t, triggered, "Should not trigger on missing branch filter")

	triggered = cleanupTrigger("preview-test", &dx.Cleanup{
		Branch: "preview-test",
		Event:  dx.BranchDeleted,
	})
	assert.False(t, triggered, "Should not trigger on missing app")
}

func Test_kustomizationTemplateAndWrite(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	m := &dx.Manifest{
		Env: "staging",
		App: "myapp",
	}
	repoName := "test/test-app"
	repoPerEnv := false

	kustomization, err := kustomizationTemplate(m, repoName, dirToWrite, repoPerEnv)
	assert.Nil(t, err)
	assert.True(t, kustomization != nil)
	assert.Equal(t, "staging/flux/kustomization-myapp.yaml", kustomization.Path)

	repoPerEnv = true
	kustomization, err = kustomizationTemplate(m, repoName, dirToWrite, repoPerEnv)
	assert.Nil(t, err)
	assert.True(t, kustomization != nil)
	assert.Equal(t, "flux/kustomization-myapp.yaml", kustomization.Path)
}

func Test_uniqueKustomizationName(t *testing.T) {
	singleEnv := false
	owner := "gimlet-io"
	repoName := "gitops-staging-infra"
	env := "staging"
	namespace := "my-team"
	appName := "myapp"
	uniqueName := uniqueKustomizationName(singleEnv, owner, repoName, env, namespace, appName)
	assert.Equal(t, "gimlet-io-staging-infra-staging-my-team-myapp", uniqueName)

	singleEnv = true
	uniqueName = uniqueKustomizationName(singleEnv, owner, repoName, env, namespace, appName)
	assert.Equal(t, "gimlet-io-staging-infra-my-team-myapp", uniqueName)
}

func Test_empty_deploy_trigger(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "xxx",
			},
		},
		&dx.Deploy{})
	assert.False(t, triggered, "Empty trigger should not trigger")
}

func Test_commit_message_pattern_triggers(t *testing.T) {
	triggered := deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "[DEPLOY: ALL] Fixing something major",
			},
		},
		&dx.Deploy{
			CommitMessagePatterns: []string{
				"[DEPLOY: myapp-1]*",
			},
		})
	assert.True(t, triggered, "Deploy all commit message pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "[DEPLOY: ALL] Fixing something major",
			},
		},
		nil)
	assert.False(t, triggered, "Deploy all should only trigger if manifest has some deploy policy. No deploy policy means no automatic deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "feature-branch",
				Event:   *dx.PushPtr(),
				Message: "[DEPLOY: ALL] Fixing something major",
			},
		},
		&dx.Deploy{
			Branch: "main",
			Event:  dx.PushPtr(),
		})
	assert.False(t, triggered, "Mismatched branch should not trigger even with deploy all. Most likely a user error")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "[DEPLOY: myapp-1] Bugfix 123",
			},
		},
		&dx.Deploy{
			CommitMessagePatterns: []string{
				"[DEPLOY: myapp-1]*",
			},
		})
	assert.True(t, triggered, "Matching commit message pattern should trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "Bugfix 123",
			},
		},
		&dx.Deploy{
			Branch: "main",
			Event:  dx.PushPtr(),
			CommitMessagePatterns: []string{
				"[DEPLOY: myapp-1]*",
			},
		})
	assert.False(t, triggered, "Non matching commit message pattern should not trigger a deploy")

	triggered = deployTrigger(
		&dx.Artifact{
			Version: dx.Version{
				Branch:  "main",
				Event:   *dx.PushPtr(),
				Message: "[DEPLOY: myapp-1] Bugfix 123",
			},
		},
		&dx.Deploy{
			Branch: "main",
			Event:  dx.PushPtr(),
			CommitMessagePatterns: []string{
				"[DEPLOY: myapp-2]*",
			},
		})
	assert.False(t, triggered, "Non matching commit message pattern should not trigger a deploy")
}
