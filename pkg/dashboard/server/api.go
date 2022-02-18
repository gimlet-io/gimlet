package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
)

func user(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(userString)
}

func envs(w http.ResponseWriter, r *http.Request) {
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)

	envs := []*api.Env{
		// {
		// 	Name:   "staging",
		// 	Stacks: []*api.Stack{},
		// },
	}
	for _, a := range agentHub.Agents {
		for _, stack := range a.Stacks {
			stack.Env = a.Name
		}
		envs = append(envs, &api.Env{
			Name:   a.Name,
			Stacks: a.Stacks,
		})
	}

	err := decorateDeployments(r.Context(), envs)
	if err != nil {
		logrus.Errorf("cannot decorate deployments: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	db := r.Context().Value("store").(*store.Store)
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get all environments from database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	allEnvs := map[string]interface{}{}
	allEnvs["envs"] = envs
	allEnvs["envsFromDB"] = envsFromDB

	allEnvsString, err := json.Marshal(allEnvs)
	if err != nil {
		logrus.Errorf("cannot serialize envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(allEnvsString)

	time.Sleep(50 * time.Millisecond) // there is a race condition in local dev: the refetch arrives sooner
	go agentHub.ForceStateSend()
}

func agents(w http.ResponseWriter, r *http.Request) {
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)

	agents := []string{} //[]string{"staging"}
	for _, a := range agentHub.Agents {
		agents = append(agents, a.Name)
	}

	agentsString, err := json.Marshal(map[string]interface{}{
		"agents": agents,
	})
	if err != nil {
		logrus.Errorf("cannot serialize agents: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(agentsString)
}

func decorateDeployments(ctx context.Context, envs []*api.Env) error {
	dao := ctx.Value("store").(*store.Store)
	gitServiceImpl := ctx.Value("gitService").(customScm.CustomGitService)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	for _, env := range envs {
		for _, stack := range env.Stacks {
			if stack.Deployment == nil {
				continue
			}
			_, err := decorateDeploymentWithSCMData(stack.Repo, stack.Deployment, dao, gitServiceImpl, token)
			if err != nil {
				return fmt.Errorf("cannot decorate commits: %s", err)
			}
		}
	}
	return nil
}

func switchToBranch(repo *git.Repository, branch string) error {
	b := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	return worktree.Checkout(&git.CheckoutOptions{Create: false, Force: false, Branch: b})
}

func chartSchema(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	config := ctx.Value("config").(*config.Config)
	goScm := genericScm.NewGoScmHelper(config, nil)

	repo := config.Chart.Repo
	path := config.Chart.Path
	valuesSchemaPath := fmt.Sprintf("%s/%s", path, "values.schema.json")
	helmUISchemaPath := fmt.Sprintf("%s/%s", path, "helm-ui.json")

	schemaString, _, err := goScm.Content(token, repo, valuesSchemaPath, "HEAD")
	if err != nil {
		logrus.Errorf("cannot fetch schema from github: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	schemaUIString, _, err := goScm.Content(token, repo, helmUISchemaPath, "HEAD")
	if err != nil {
		logrus.Errorf("cannot fetch UI schema from github: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var schema interface{}
	err = json.Unmarshal([]byte(schemaString), &schema)
	if err != nil {
		logrus.Errorf("cannot parse schema: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var schemaUI interface{}
	err = json.Unmarshal([]byte(schemaUIString), &schemaUI)
	if err != nil {
		logrus.Errorf("cannot parse UI schema: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	schemas := map[string]interface{}{}
	schemas["chartSchema"] = schema
	schemas["uiSchema"] = schemaUI

	schemasString, err := json.Marshal(schemas)
	if err != nil {
		logrus.Errorf("cannot serialize schemas: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(schemasString))
}

func application(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gitServiceImpl := ctx.Value("gitService").(customScm.CustomGitService)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)

	tokenString, err := tokenManager.AppToken()
	if err != nil {
		logrus.Errorf("cannot generate application token: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	appName, appSettingsURL, installationURL, err := gitServiceImpl.GetAppNameAndAppSettingsURLs(tokenString, ctx)
	if err != nil {
		logrus.Errorf("cannot get app info: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	appinfos := map[string]interface{}{}
	appinfos["appName"] = appName
	appinfos["installationURL"] = installationURL
	appinfos["appSettingsURL"] = appSettingsURL

	appinfosString, err := json.Marshal(appinfos)
	if err != nil {
		logrus.Errorf("cannot serialize appinfos: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(appinfosString))

}

func saveEnvToDB(w http.ResponseWriter, r *http.Request) {
	var envNameToSave string
	err := json.NewDecoder(r.Body).Decode(&envNameToSave)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	db := r.Context().Value("store").(*store.Store)
	envToSave := &model.Environment{
		Name: envNameToSave,
	}
	err = db.CreateEnvironment(envToSave)
	if err != nil {
		logrus.Errorf("cannot create environment to database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(envNameToSave))
}

func deleteEnvFromDB(w http.ResponseWriter, r *http.Request) {
	var envNameToDelete string
	err := json.NewDecoder(r.Body).Decode(&envNameToDelete)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	db := r.Context().Value("store").(*store.Store)

	err = db.DeleteEnvironment(envNameToDelete)
	if err != nil {
		logrus.Errorf("cannot delete environment to database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(envNameToDelete))
}
