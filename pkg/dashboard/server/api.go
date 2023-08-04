package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
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
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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

func envs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	envs := []*api.GitopsEnv{}
	for _, env := range envsFromDB {
		repo, err := gitRepoCache.InstanceForRead(env.InfraRepo)
		if err != nil {
			if strings.Contains(err.Error(), "repository not found") ||
				strings.Contains(err.Error(), "repo name is mandatory") {
				envs = append(envs, &api.GitopsEnv{
					Name: env.Name,
				})
				continue
			} else {
				logrus.Errorf("cannot get repo: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		var stackConfig *dx.StackConfig
		if env.RepoPerEnv {
			stackConfig, err = stackYaml(repo, "stack.yaml")
			if err != nil && !strings.Contains(err.Error(), "file not found") {
				logrus.Errorf("cannot get stack yaml from repo: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			stackConfig, err = stackYaml(repo, filepath.Join(env.Name, "stack.yaml"))
			if err != nil && !strings.Contains(err.Error(), "file not found") {
				logrus.Errorf("cannot get stack yaml from %s repo for env %s: %s", env.InfraRepo, env.Name, err)
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

		envs = append(envs, &api.GitopsEnv{
			Name:                        env.Name,
			InfraRepo:                   env.InfraRepo,
			AppsRepo:                    env.AppsRepo,
			RepoPerEnv:                  env.RepoPerEnv,
			KustomizationPerApp:         env.KustomizationPerApp,
			BuiltIn:                     env.BuiltIn,
			StackConfig:                 stackConfig,
			StackDefinition:             stackDefinition,
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

	time.Sleep(50 * time.Millisecond) // there is a race condition in local dev: the refetch arrives sooner
	go agentHub.ForceStateSend()
}

func getPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	serviceName := r.URL.Query().Get("serviceName")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.StreamPodLogsSend(namespace, serviceName)
}

func stopPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	serviceName := r.URL.Query().Get("serviceName")

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agentHub.StopPodLogs(namespace, serviceName)
}

func getAlerts(w http.ResponseWriter, r *http.Request) {
	// db := r.Context().Value("store").(*store.Store)
	// alerts, err := db.FiringAlerts()
	// if err != nil {
	// 	logrus.Errorf("cannot get alerts from database: %s", err)
	// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	// }

	alertsString, err := json.Marshal([]*model.Alert{}) //(alerts)
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
			_, err := decorateDeploymentWithSCMData(stack.Repo, stack.Deployment, dao, gitServiceImpl, token)
			if err != nil {
				return fmt.Errorf("cannot decorate commits: %s", err)
			}
		}
	}
	return nil
}

func chartSchema(w http.ResponseWriter, r *http.Request) {
	var chartFromBody dx.Chart
	err := json.NewDecoder(r.Body).Decode(&chartFromBody)
	if err != nil {
		logrus.Errorf("cannot decode env name to save: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	dashConfig := ctx.Value("config").(*config.Config)
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	env := chi.URLParam(r, "env")
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	installationToken, _, _ := tokenManager.Token()

	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	validChart := dashConfig.Chart
	if chartFromBody.Name != "" {
		validChart = config.Chart{
			Name:    chartFromBody.Name,
			Repo:    chartFromBody.Repository,
			Version: chartFromBody.Version,
		}
	}

	m, err := getManifest(&validChart, repo, env)
	if err != nil {
		logrus.Errorf("cannot get manifest: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	schemaString, schemaUIString, err := dx.ChartSchema(m, installationToken)
	if err != nil {
		logrus.Errorf("cannot get schema from manifest: %s", err)
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

	chartReference := chartFromConfig(&validChart)

	schemas := map[string]interface{}{}
	schemas["schema"] = schema
	schemas["uiSchema"] = schemaUI
	schemas["reference"] = chartReference

	schemasString, err := json.Marshal(schemas)
	if err != nil {
		logrus.Errorf("cannot serialize schemas: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(schemasString))
}

func charts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)

	parsedCharts, err := parseCharts(config.Charts)
	if err != nil {
		logrus.Errorf("cannot parse charts from config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var charts []ConfigChart
	for _, chart := range parsedCharts {
		charts = append(charts, ConfigChart{
			Name:      chart.Name,
			Reference: chartFromConfig(chart),
		})
	}

	chartsString, err := json.Marshal(charts)
	if err != nil {
		logrus.Errorf("cannot serialize charts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chartsString))
}

func parseCharts(chartsString string) ([]*config.Chart, error) {
	charts := []*config.Chart{}
	splittedCharts := strings.Split(chartsString, ";")

	for _, chartsString := range splittedCharts {
		if chartsString == "" {
			continue
		}

		parsedchartsString, err := parse(chartsString)
		if err != nil {
			return nil, fmt.Errorf("invalid charts format: %s", err)
		}

		chart := &config.Chart{
			Name:    parsedchartsString.Get("name"),
			Repo:    parsedchartsString.Get("repo"),
			Version: parsedchartsString.Get("version"),
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

func parse(query string) (url.Values, error) {
	m := make(url.Values)
	err := parseQuery(m, query)
	return m, err
}

func parseQuery(m url.Values, query string) (err error) {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, ",")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return err
}

func getManifest(chart *config.Chart, repo *git.Repository, env string) (*dx.Manifest, error) {
	defaultManifest := &dx.Manifest{
		Chart: chartFromConfig(chart),
	}

	branch, err := helper.HeadBranch(repo)
	if err != nil {
		return nil, err
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, branch, ".gimlet")
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			return defaultManifest, nil
		} else {
			return nil, err
		}
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
		if envConfig.Env == env {
			return &envConfig, nil
		}
	}

	return defaultManifest, nil
}

func chartFromConfig(chart *config.Chart) dx.Chart {
	if strings.HasPrefix(chart.Name, "git@") ||
		strings.Contains(chart.Name, ".git") {
		return dx.Chart{
			Name: chart.Name,
		}
	}

	return dx.Chart{
		Repository: chart.Repo,
		Name:       chart.Name,
		Version:    chart.Version,
	}
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
		true,
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
