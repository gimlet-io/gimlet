package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"path/filepath"

	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitnami-labs/sealed-secrets/pkg/crypto"
	"github.com/gimlet-io/capacitor/pkg/flux"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/alert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	commitHelper "github.com/gimlet-io/gimlet-cli/pkg/dashboard/commits"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/gimlet-io/gimlet-cli/pkg/version"
	"github.com/go-chi/chi"
	"github.com/go-git/go-git/v5"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/cert"
	"sigs.k8s.io/yaml"
)

const fluxPattern = "flux-%s"

func user(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)

	token := token.New(token.UserToken, user.Login)
	tokenStr, err := token.Sign(user.Secret)
	if err != nil {
		logrus.Errorf("couldn't generate JWT token %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// token is not saved as it is JWT
	user.Token = tokenStr

	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(userString)
}

func saveUser(w http.ResponseWriter, r *http.Request) {
	var usernameToSave string
	err := json.NewDecoder(r.Body).Decode(&usernameToSave)
	if err != nil {
		logrus.Errorf("cannot decode user name to save: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	user := &model.User{
		Login:  usernameToSave,
		Secret: base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
	}

	err = store.CreateUser(user)
	if err != nil {
		logrus.Errorf("cannot creat user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	token := token.New(token.UserToken, user.Login)
	tokenStr, err := token.Sign(user.Secret)
	if err != nil {
		logrus.Errorf("couldn't create user token %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// token is not saved as it is JWT
	user.Token = tokenStr

	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(userString)
}

func envs(w http.ResponseWriter, r *http.Request) {
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)

	connectedAgents := []*api.ConnectedAgent{}
	for _, a := range agentHub.Agents {
		for _, stack := range a.Stacks {
			stack.Env = a.Name
		}
		connectedAgents = append(connectedAgents, &api.ConnectedAgent{
			Name:      a.Name,
			Stacks:    a.Stacks,
			FluxState: a.FluxState,
		})
	}

	err := decorateDeployments(r.Context(), connectedAgents)
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

	envs := []*api.GitopsEnv{}
	for _, env := range envsFromDB {
		envs = append(envs, &api.GitopsEnv{
			Name:                        env.Name,
			InfraRepo:                   env.InfraRepo,
			AppsRepo:                    env.AppsRepo,
			RepoPerEnv:                  env.RepoPerEnv,
			KustomizationPerApp:         env.KustomizationPerApp,
			BuiltIn:                     env.BuiltIn,
			DeploymentAutomationEnabled: false,
		})
	}

	for _, env := range envs {
		env.DeploymentAutomationEnabled = deploymentAutomationEnabled(env.Name, envs)
	}

	allEnvs := map[string]interface{}{}
	allEnvs["connectedAgents"] = connectedAgents
	allEnvs["envs"] = envs

	allEnvsString, err := json.Marshal(allEnvs)
	if err != nil {
		logrus.Errorf("cannot serialize envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(allEnvsString)

	go agentHub.ForceStateSend()
}

func fluxStateHandler(w http.ResponseWriter, r *http.Request) {
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)

	fluxStates := map[string]*flux.FluxState{}
	for _, a := range agentHub.Agents {
		fluxStates[a.Name] = a.FluxStatev2
	}

	fluxStatesString, err := json.Marshal(fluxStates)
	if err != nil {
		logrus.Errorf("cannot serialize envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(fluxStatesString)
}

func getPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	deployment := r.URL.Query().Get("deploymentName")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.StreamPodLogsSend(namespace, deployment)
}

func stopPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	deployment := r.URL.Query().Get("deploymentName")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.StopPodLogs(namespace, deployment)
}

func getDeploymentDetails(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	deployment := r.URL.Query().Get("name")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.DeploymentDetails(namespace, deployment)
}

func getPodDetails(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.PodDetails(namespace, name)
}

func reconcile(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.ReconcileResource(resource, namespace, name)
}

func getAlerts(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("store").(*store.Store)
	dbAlerts, err := db.Alerts()
	if err != nil {
		logrus.Errorf("cannot get alerts from database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	thresholds := alert.Thresholds()
	decoratedAlerts := []*api.Alert{}
	for _, dbAlert := range dbAlerts {
		silencedUntil, err := db.DeploymentSilencedUntil(dbAlert.DeploymentName, dbAlert.Type)
		if err != nil {
			logrus.Errorf("couldn't get deployment silenced until: %s", err)
		}

		t := alert.ThresholdByType(thresholds, dbAlert.Type)
		decoratedAlerts = append(decoratedAlerts, api.NewAlert(dbAlert, t.Text(), t.Name(), silencedUntil))
	}

	alertsString, err := json.Marshal(decoratedAlerts)
	if err != nil {
		logrus.Errorf("cannot serialize alerts: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(alertsString)
}

func deploymentAutomationEnabled(envName string, envs []*api.GitopsEnv) bool {
	for _, env := range envs {
		if env.StackConfig == nil {
			continue
		}

		if _, ok := env.StackConfig.Config["gimletd"]; !ok {
			continue
		}

		gimletdConfig := env.StackConfig.Config["gimletd"].(map[string]interface{})
		if gimletdEnvs, ok := gimletdConfig["environments"]; ok {
			for _, e := range gimletdEnvs.([]interface{}) {
				envConfig := e.(map[string]interface{})
				if envConfig["name"] == envName {
					return true
				}
			}
		}
	}

	return false
}

func stackConfig(w http.ResponseWriter, r *http.Request) {
	envName := chi.URLParam(r, "env")
	db := r.Context().Value("store").(*store.Store)

	env, err := db.GetEnvironment(envName)
	if err != nil {
		logrus.Errorf("cannot get env: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var stackConfig *dx.StackConfig
	stackYamlPath := "stack.yaml"
	if !env.RepoPerEnv {
		stackYamlPath = filepath.Join(env.Name, "stack.yaml")
	}

	gitRepoCache, _ := r.Context().Value("gitRepoCache").(*nativeGit.RepoCache)
	err = gitRepoCache.PerformAction(env.InfraRepo, func(repo *git.Repository) error {
		var inerErr error
		stackConfig, inerErr = stackYaml(repo, stackYamlPath)
		return inerErr
	})
	if err != nil {
		if !strings.Contains(err.Error(), "file not found") {
			logrus.Errorf("cannot get stack yaml from repo: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else {
			logrus.Errorf("cannot get repo: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	stackDefinition, err := loadStackDefinition(stackConfig)
	if err != nil && !strings.Contains(err.Error(), "file not found") {
		logrus.Errorf("cannot get stack definition: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitopsEnv := &api.GitopsEnv{
		Name:            env.Name,
		StackConfig:     stackConfig,
		StackDefinition: stackDefinition,
	}

	gitopsEnvString, err := json.Marshal(gitopsEnv)
	if err != nil {
		logrus.Errorf("cannot serialize envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(gitopsEnvString)
}

func loadStackDefinition(stackConfig *dx.StackConfig) (map[string]interface{}, error) {
	var url string
	if stackConfig != nil {
		url = stackConfig.Stack.Repository
	} else {
		latestTag, _ := stack.LatestVersion(stack.DefaultStackURL)
		if latestTag != "" {
			url = stack.DefaultStackURL + "?tag=" + latestTag
		}
	}

	stackDefinitionYaml, err := stack.StackDefinitionFromRepo(url)
	if err != nil {
		return nil, fmt.Errorf("cannot get stack definition: %s", err.Error())
	}

	var stackDefinition map[string]interface{}
	err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
	return stackDefinition, err
}

func stackYaml(repo *git.Repository, path string) (*dx.StackConfig, error) {
	var stackConfig dx.StackConfig

	headBranch, err := helper.HeadBranch(repo)
	if err != nil {
		return nil, err
	}

	yamlString, err := helper.RemoteContentOnBranchWithoutCheckout(repo, headBranch, path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(yamlString), &stackConfig)
	if err != nil {
		return nil, err
	}

	return &stackConfig, nil
}

func agents(w http.ResponseWriter, r *http.Request) {
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)

	agents := []string{}
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

func decorateDeployments(ctx context.Context, envs []*api.ConnectedAgent) error {
	dao := ctx.Value("store").(*store.Store)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	for _, env := range envs {
		for _, stack := range env.Stacks {
			if stack.Deployment == nil {
				continue
			}

			err := commitHelper.AssureSCMData(
				stack.Repo,
				[]string{stack.Deployment.SHA},
				dao,
				gitServiceImpl,
				token,
			)
			if err != nil {
				return fmt.Errorf("cannot assure commit data: %s", err)
			}

			dbCommits, err := dao.CommitsByRepoAndSHA(stack.Repo, []string{stack.Deployment.SHA})
			if err != nil {
				return fmt.Errorf("cannot get commits from db: %s", err)
			}

			if len(dbCommits) > 0 {
				stack.Deployment.CommitMessage = dbCommits[0].Message
			}
		}
	}
	return nil
}

func defaultDeploymentTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	installationToken, _, _ := tokenManager.Token()

	templates, err := deploymentTemplates(config.Charts, installationToken)
	if err != nil {
		logrus.Errorf("cannot convert charts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	templatesString, err := json.Marshal(templates)
	if err != nil {
		logrus.Errorf("cannot serialize charts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(templatesString))
}

func deploymentTemplateForApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	env := chi.URLParam(r, "env")
	configName := chi.URLParam(r, "config")
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	installationToken, _, _ := tokenManager.Token()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	var appChart *dx.Chart
	var err error
	gitRepoCache.PerformAction(fmt.Sprintf("%s/%s", owner, repoName),
		func(repo *git.Repository) error {
			var innerErr error
			appChart, innerErr = getChartForApp(repo, env, configName)
			return innerErr
		})
	if err != nil {
		logrus.Errorf("cannot get manifest chart: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	templates, err := deploymentTemplates([]dx.Chart{*appChart}, installationToken)
	if err != nil {
		logrus.Errorf("cannot convert charts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	templatesString, err := json.Marshal(templates)
	if err != nil {
		logrus.Errorf("cannot serialize charts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(templatesString))
}

func deploymentTemplates(charts []dx.Chart, installationToken string) ([]DeploymentTemplate, error) {
	var templates []DeploymentTemplate
	for _, chart := range charts {
		m := &dx.Manifest{
			Chart: chart,
		}

		schemaString, schemaUIString, err := dx.ChartSchema(m, installationToken)
		if err != nil {
			return nil, err
		}

		var schema interface{}
		err = json.Unmarshal([]byte(schemaString), &schema)
		if err != nil {
			return nil, err
		}

		var schemaUI interface{}
		err = json.Unmarshal([]byte(schemaUIString), &schemaUI)
		if err != nil {
			return nil, err
		}

		templates = append(templates, DeploymentTemplate{
			Reference: chart,
			Schema:    schema,
			UISchema:  schemaUI,
		})
	}
	return templates, nil
}

func getChartForApp(repo *git.Repository, env string, app string) (*dx.Chart, error) {
	branch, err := helper.HeadBranch(repo)
	if err != nil {
		return nil, err
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, branch, ".gimlet")
	if err != nil {
		return nil, err
	}

	envConfigs := []dx.Manifest{}
	for _, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		envConfigs = append(envConfigs, envConfig)
	}

	for _, envConfig := range envConfigs {
		if envConfig.Env == env && envConfig.App == app {
			return &envConfig.Chart, nil
		}
	}

	return nil, nil
}

func application(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
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
	appinfos["dashboardVersion"] = version.String()

	appinfosString, err := json.Marshal(appinfos)
	if err != nil {
		logrus.Errorf("cannot serialize appinfos: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(appinfosString))
}

func seal(w http.ResponseWriter, r *http.Request) {
	var secret string
	err := json.NewDecoder(r.Body).Decode(&secret)
	if err != nil {
		logrus.Errorf("cannot decode secret: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	env := chi.URLParam(r, "env")
	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	cert, err := extractCert(agentHub.Agents, env)
	if err != nil {
		logrus.Errorf("cannot extract certificate from agenthub: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	key, err := parseKey(cert)
	if err != nil {
		logrus.Errorf("cannot parse public key: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	sealedValue, err := sealValue(key, secret)
	if err != nil {
		logrus.Errorf("cannot seal item: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sealedValue))
}

func extractCert(agents map[string]*streaming.ConnectedAgent, env string) ([]byte, error) {
	if agent, ok := agents[env]; ok {
		if len(agent.Certificate) != 0 {
			return agent.Certificate, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func parseKey(data []byte) (*rsa.PublicKey, error) {
	certs, err := cert.ParseCertsPEM(data)
	if err != nil {
		return nil, err
	}

	cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected RSA public key but found %v", certs[0].PublicKey)
	}

	return cert, nil
}

func sealValue(pubKey *rsa.PublicKey, data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("empty secret")
	}

	clusterWide := []byte("")
	result, err := crypto.HybridEncrypt(rand.Reader, pubKey, []byte(data), clusterWide)
	return base64.StdEncoding.EncodeToString(result), err
}

func silenceAlert(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Query().Get("object")
	until := r.URL.Query().Get("until")

	db := r.Context().Value("store").(*store.Store)
	err := db.SaveKeyValue(&model.KeyValue{
		Key:   object,
		Value: until,
	})
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func restartDeployment(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.RestartDeployment(namespace, name)
}

func saveEnvToDB(w http.ResponseWriter, r *http.Request) {
	var envNameToSave string
	err := json.NewDecoder(r.Body).Decode(&envNameToSave)
	if err != nil {
		logrus.Errorf("cannot decode env name to save: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	lowerCaseEnvNameToSave := strings.ToLower(envNameToSave)
	db := r.Context().Value("store").(*store.Store)
	envToSave := &model.Environment{
		Name: lowerCaseEnvNameToSave,
	}
	err = db.CreateEnvironment(envToSave)
	if err != nil {
		logrus.Errorf("cannot create environment to database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(lowerCaseEnvNameToSave))
}

func spinOutBuiltInEnv(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	_, err := store.KeyValue(model.SpinnedOut)
	if err == nil {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed)+" - built-in env already spinned out", http.StatusPreconditionFailed)
		return
	}

	envs, err := store.GetEnvironments()
	if err != nil {
		logrus.Errorf("cannot get envs: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	var builtInEnv *model.Environment
	for _, env := range envs {
		if env.BuiltIn {
			builtInEnv = env
			break
		}
	}
	if builtInEnv == nil {
		http.Error(w, http.StatusText(http.StatusPreconditionFailed)+" - built-in environment missing", http.StatusPreconditionFailed)
		return
	}

	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
	gitToken, _, _ := tokenManager.Token()

	user := ctx.Value("user").(*model.User)

	oldInfraRepo := builtInEnv.InfraRepo
	oldAppsRepo := builtInEnv.AppsRepo
	builtInEnv.InfraRepo = fmt.Sprintf("%s/gitops-%s-infra", dynamicConfig.Org(), builtInEnv.Name)
	builtInEnv.AppsRepo = fmt.Sprintf("%s/gitops-%s-apps", dynamicConfig.Org(), builtInEnv.Name)

	// Creating repos
	_, err = AssureRepoExists(
		builtInEnv.InfraRepo,
		user.AccessToken,
		gitToken,
		user.Login,
		gitServiceImpl,
	)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, err = AssureRepoExists(
		builtInEnv.AppsRepo,
		user.AccessToken,
		gitToken,
		user.Login,
		gitServiceImpl,
	)
	if err != nil {
		logrus.Errorf("cannot assure repo exists: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	gitUser, err := store.User("git")
	if err != nil {
		logrus.Errorf("cannot get git user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _, err = MigrateEnv(
		gitRepoCache,
		gitServiceImpl,
		builtInEnv.Name,
		oldInfraRepo,
		builtInEnv.InfraRepo,
		true,
		gitToken,
		true,
		true,
		false,
		false,
		dynamicConfig.ScmURL(),
		gitUser,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _, err = MigrateEnv(
		gitRepoCache,
		gitServiceImpl,
		builtInEnv.Name,
		oldAppsRepo,
		builtInEnv.AppsRepo,
		true,
		gitToken,
		false,
		false,
		false,
		false,
		dynamicConfig.ScmURL(),
		gitUser,
	)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// preventing from creating another built-in env
	err = store.SaveKeyValue(&model.KeyValue{
		Key:   model.SpinnedOut,
		Value: "true",
	})
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	builtInEnv.BuiltIn = false
	err = store.UpdateEnvironment(builtInEnv)
	if err != nil {
		logrus.Errorf("cannot update env: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resultBytes, err := json.Marshal(builtInEnv)
	if err != nil {
		log.Errorf("could not serialize results: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultBytes)
}

func deleteEnvFromDB(w http.ResponseWriter, r *http.Request) {
	var envNameToDelete string
	err := json.NewDecoder(r.Body).Decode(&envNameToDelete)
	if err != nil {
		logrus.Errorf("cannot decode env name to delete: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	db := r.Context().Value("store").(*store.Store)
	err = db.DeleteEnvironment(envNameToDelete)
	if err != nil {
		logrus.Errorf("cannot delete environment to database: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fluxUser := fmt.Sprintf(fluxPattern, envNameToDelete)
	err = db.DeleteUser(fluxUser)
	if err != nil {
		logrus.Errorf("cannot delete user %s: %s", fluxUser, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(envNameToDelete))
}

func getFlags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	var provider string
	termsOfServiceFeatureFlag := config.TermsOfServiceFeatureFlag

	if dynamicConfig.IsGithub() {
		provider = "GitHub"
	} else if dynamicConfig.IsGitlab() {
		provider = "GitLab"
	}

	data := map[string]interface{}{
		"provider":                  provider,
		"termsOfServiceFeatureFlag": termsOfServiceFeatureFlag,
	}

	dataString, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("cannot serialize data: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(dataString))
}
