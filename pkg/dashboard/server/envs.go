package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	dNativeGit "github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type saveInfrastructureComponentsReq struct {
	Env                      string                 `json:"env"`
	IsPerRepository          bool                   `json:"isPerRepository"`
	InfrastructureComponents map[string]interface{} `json:"infrastructureComponents"`
}

func saveInfrastructureComponents(w http.ResponseWriter, r *http.Request) {
	var req saveInfrastructureComponentsReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logrus.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	org := config.Github.Org

	repoName := fmt.Sprintf("%s/gitops-infra", org)
	if req.IsPerRepository {
		repoName = fmt.Sprintf("%s/gitops-%s-infra", org, req.Env)
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*dNativeGit.RepoCache)
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoName)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var stackConfig *dx.StackConfig
	stackYamlPath := filepath.Join(req.Env, "stack.yaml")
	if req.IsPerRepository {
		stackYamlPath = "stack.yaml"
	}

	stackConfig, err = stackYaml(repo, stackYamlPath)
	if err != nil {
		logrus.Errorf("cannot get stack yaml from repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stackConfig.Config = req.InfrastructureComponents
	stackConfigBuff := bytes.NewBufferString("")
	e := yaml.NewEncoder(stackConfigBuff)
	e.SetIndent(2)
	err = e.Encode(stackConfig)
	if err != nil {
		logrus.Errorf("cannot serialize stack config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), dNativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		logrus.Errorf("cannot get working copy: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		logrus.Errorf("cannot stage changes: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, err = nativeGit.Commit(repo, "[Gimlet Dashboard] Updating infrastructure components")
	if err != nil {
		logrus.Errorf("cannot commit changes: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	err = nativeGit.PushWithToken(repo, token)
	if err != nil {
		logrus.Errorf("cannot push changes: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache.Invalidate(repoName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func stackDefinition(w http.ResponseWriter, r *http.Request) {
	// TODO get it from Github
	stackDefinitionString := `
	{
		"name": "Gimlet Stack Reference",
		"description": "Providing a reference stack to demonstrate Gimlet Stack",
		"categories": [
		  {
			"name": "⬅️Ingress",
			"id": "ingress"
		  },
		  {
			"name": "\uD83D\uDCD1 Logging",
			"id": "logging"
		  }
		],
		"components": [
		  {
			"name": "Nginx",
			"variable": "nginx",
			"category": "ingress",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/nginx.png",
			"description": "",
			"onePager": "An Nginx proxy server that routes traffic to your applications based on the host name or path.",
			"schema": {
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "$id": "#nginx",
			  "type": "object",
			  "properties": {
				"enabled": {
				  "$id": "#/properties/enabled",
				  "type": "boolean",
				  "title": "Enabled"
				}
			  },
			  "dependencies": {
				"enabled": {
				  "oneOf": [
					{
					  "properties": {
						"enabled": {
						  "const": false
						}
					  }
					},
					{
					  "properties": {
						"enabled": {
						  "const": true
						},
						"host": {
						  "$id": "#/properties/host",
						  "type": "string",
						  "title": "Host",
						  "description": "Your company domain you will expose your services on"
						}
					  },
					  "required": [
						"host"
					  ]
					}
				  ]
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#nginx"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  },
		  {
			"name": "CertManager",
			"variable": "certManager",
			"category": "ingress",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/certManager.png",
			"description": "",
			"onePager": "",
			"schema": {
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "$id": "http://example.com/example.json",
			  "type": "object",
			  "title": "The root schema",
			  "description": "The root schema comprises the entire JSON document.",
			  "required": [
				"email"
			  ],
			  "properties": {
				"enabled": {
				  "$id": "#/properties/enabled",
				  "type": "boolean",
				  "title": "Enabled"
				},
				"email": {
				  "$id": "#/properties/email",
				  "type": "string",
				  "title": "Administrator email",
				  "description": "Let's Encrypt will email you on this email upon expiring certificates",
				  "default": "",
				  "examples": [
					"admin@mycompany.com"
				  ]
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#/properties/enabled",
				  "#/properties/email"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  },
		  {
			"name": "Loki",
			"variable": "loki",
			"category": "logging",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/loki.png",
			"description": "",
			"onePager": "",
			"schema": {
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "$id": "http://example.com/example.json",
			  "type": "object",
			  "title": "The root schema",
			  "description": "The root schema comprises the entire JSON document.",
			  "properties": {
				"enabled": {
				  "$id": "#/properties/enabled",
				  "type": "boolean",
				  "title": "Enabled"
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#/properties/enabled"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  }
		]
	  }
	  `

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(stackDefinitionString))
}

func bootstrapGitops(w http.ResponseWriter, r *http.Request) {
	bootstrapConfig := &api.GitopsBootstrapConfig{}
	err := json.NewDecoder(r.Body).Decode(&bootstrapConfig)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	org := config.Github.Org
	var gitopsRepoURL string
	var env string
	var repoName string
	var repoPath string

	if bootstrapConfig.RepoPerEnv {
		repoName = fmt.Sprintf("gitops-%s-infra", bootstrapConfig.EnvName)
		repoPath = fmt.Sprintf("%s/%s", org, repoName)
		gitopsRepoURL = fmt.Sprintf("git@github.com:%s.git", repoPath)
	} else {
		repoName = "gitops-infra"
		repoPath = fmt.Sprintf("%s/%s", org, repoName)
		env = bootstrapConfig.EnvName
		gitopsRepoURL = fmt.Sprintf("git@github.com:%s.git", repoPath)
	}

	err = assureRepoExists(ctx, repoPath, repoName, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*dNativeGit.RepoCache)
	repo, tmpPath, err := gitRepoCache.InstanceForWrite(repoPath)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	_, _, _, err = gitops.GenerateManifests(
		true,
		env,
		bootstrapConfig.RepoPerEnv,
		tmpPath,
		true,
		true,
		gitopsRepoURL,
		"main",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	folderToAdd := folderToAdd(bootstrapConfig.EnvName, bootstrapConfig.RepoPerEnv)
	err = commitAndPush(repo, token, folderToAdd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	gitRepoCache.Invalidate(repoName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func assureRepoExists(ctx context.Context, repoPath string, repoName string, token string) error {
	orgRepos, err := getOrgRepos(ctx)
	if err != nil {
		return err
	}

	if !hasRepo(orgRepos, repoPath) {
		personalToken := ""

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: personalToken})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		var (
			name     = repoName
			private  = true
			autoInit = true
		)

		r := &github.Repository{
			Name:     &name,
			Private:  &private,
			AutoInit: &autoInit}
		_, _, err = client.Repositories.Create(ctx, "", r)
		if err != nil {
			return err
		}
	}

	return nil
}

func commitAndPush(repo *git.Repository, token string, folderToAdd string) error {
	err := nativeGit.StageFolder(repo, fmt.Sprintf("./%s", folderToAdd))
	if err != nil {
		return err
	}

	_, err = nativeGit.Commit(repo, "Gimlet Bootstrapping")
	if err != nil {
		return err
	}

	err = nativeGit.PushWithToken(repo, token)
	if err != nil {
		return err
	}

	return nil
}

func folderToAdd(envName string, repoPerEnv bool) string {
	if repoPerEnv {
		return "flux"
	} else {
		return envName
	}
}
