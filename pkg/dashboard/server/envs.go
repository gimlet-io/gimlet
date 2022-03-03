package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	dNativeGit "github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func saveInfrastructureComponents(w http.ResponseWriter, r *http.Request) {
	var infrastructureComponents map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&infrastructureComponents)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	infrastructureComponentsString, err := json.Marshal(infrastructureComponents)
	if err != nil {
		logrus.Errorf("cannot serialize infrastructure components: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(infrastructureComponentsString)
}

func stackDefinition(w http.ResponseWriter, r *http.Request) {
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

func stack(w http.ResponseWriter, r *http.Request) {
	stackString := `
	{
		"certManager": {
    "email": "laszlo@gimlet.io",
    "enabled": "true"
		},
  "loki": {
    "enabled": "true",
    "persistence": "true"
  },
  "nginx": {
    "enabled": "true",
    "host": "staging.gimlet.io"
  },
  "prometheus": {
    "enabled": "true",
    "persistence": "true"
  },
  "sealedSecrets": {
    "enabled": "true"
  }
	}
	`

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(stackString))
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

	err = commitAndPush(repo, token, bootstrapConfig.RepoPerEnv, bootstrapConfig.EnvName, tmpPath)
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

	for _, orgRepo := range orgRepos {
		if orgRepo == repoPath {
			return nil
		}
	}

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

	return nil
}

func commitAndPush(repo *git.Repository, token string, repoPerEnv bool, envName string, repoPath string) error {
	var folderToAdd string
	var gitMessage string

	if repoPerEnv {
		folderToAdd = "flux"
		gitMessage = "Gimlet Bootstrapping"
	} else {
		folderToAdd = envName
		gitMessage = fmt.Sprintf("Gimlet Bootstrapping %s", envName)
	}

	err := nativeGit.StageFolder(repo, fmt.Sprintf("./%s", folderToAdd))
	if err != nil {
		return err
	}

	_, err = nativeGit.Commit(repo, gitMessage)
	if err != nil {
		return err
	}

	//GIT PULL

	err = nativeGit.PushWithToken(repo, token, repoPath, repoPerEnv)
	if err != nil {
		return err
	}

	return nil
}
