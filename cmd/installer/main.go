package main

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/installer/web"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/httputil"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"

	"encoding/hex"
	"math/rand"
)

type data struct {
	id                      string
	slug                    string
	clientId                string
	clientSecret            string
	pem                     string
	appOwner                string
	installationId          string
	tokenManager            customScm.NonImpersonatedTokenManager
	accessToken             string
	refreshToken            string
	loggedInUser            string
	repoCache               *nativeGit.RepoCache
	gimletdPublicKey        string
	infraGitopsRepoFileName string
	infraPublicKey          string
	infraSecretFileName     string
	appsGitopsRepoFileName  string
	appsPublicKey           string
	appsSecretFileName      string
	notificationsFileName   string
	infraRepo               string
	appsRepo                string
	repoPerEnv              bool
	envName                 string
	stackConfig             *dx.StackConfig
	stackDefinition         map[string]interface{}
	securityToken           string
	gitlab                  bool
	github                  bool
	scmURL                  string
}

func main() {
	fmt.Println("Installer init..")
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	stackUrl := stack.DefaultStackURL
	// latestTag, _ := stack.LatestVersion(stackUrl)
	// if latestTag != "" {
	stackUrl = stackUrl + "?branch=latest-gimlet" // + latestTag
	// }

	stackConfig := &dx.StackConfig{
		Stack: dx.StackRef{
			Repository: stackUrl,
		},
		Config: map[string]interface{}{},
	}

	stackDefinition, err := loadStackDefinition(stackConfig)
	if err != nil {
		panic(err)
	}

	r.Use(middleware.WithValue("data", &data{
		stackConfig:     stackConfig,
		stackDefinition: stackDefinition,
		scmURL:          "github.com", // will be overwritten for Gitlab
	}))

	browserClosed := make(chan int, 1)

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(browserClosed, w, r)
	})
	r.Get("/context", getContext)
	r.Get("/created", created)
	r.Get("/auth", auth)
	r.Get("/installed", installed)
	r.Post("/bootstrap", bootstrap)
	r.Post("/done", done)
	r.Post("/gitlabInit", gitlabInit)

	workDir, err := ioutil.TempDir(os.TempDir(), "gimlet")
	if err != nil {
		panic(err)
	}
	writeTempFiles(workDir)
	defer removeTempFiles(workDir)

	fs := http.FileServer(http.Dir(workDir))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(workDir + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	srv := http.Server{Addr: ":9000", Handler: r}
	go func() {
		err = srv.ListenAndServe()
		if err != nil && err.Error() != "http: Server closed" {
			panic(err)
		}
	}()

	fmt.Println("Installer started")

	select {
	case <-ctrlC:
		os.Exit(-1)
	case <-browserClosed:
	}

	srv.Shutdown(context.TODO())
}

func initStackConfig(data *data) {
	jwtSecret, _ := randomHex(32)
	agentAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)
	_, agentToken, _ := agentAuth.Encode(map[string]interface{}{"user_id": "gimlet-agent"})

	webhookSecret, _ := randomHex(32)

	postgresPassword := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	dashboardPassword := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	gimletdAdminToken := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	token := token.New(token.UserToken, "admin")
	gimletdSignedAdminToken, err := token.Sign(gimletdAdminToken)
	if err != nil {
		panic(err)
	}

	dashboardPostgresConfig := map[string]interface{}{
		"install":          true,
		"hostAndPort":      "postgresql:5432",
		"postgresPassword": postgresPassword,
		"db":               "gimlet",
		"user":             "gimlet",
		"password":         dashboardPassword,
	}

	data.stackConfig.Config["gimletAgent"] = map[string]interface{}{
		"enabled":          true,
		"environment":      "will be set from user input on the ui",
		"dashboardAddress": "http://gimlet:9000",
		"agentKey":         agentToken,
	}

	data.stackConfig.Config["gimlet"] = map[string]interface{}{
		"enabled":       true,
		"jwtSecret":     jwtSecret,
		"gimletdToken":  gimletdSignedAdminToken,
		"webhookSecret": webhookSecret,
		"bootstrapEnv":  "will be set right before bootstrap",
		"postgresql":    dashboardPostgresConfig,
		"gimletdURL":    "http://gimletd:8888",
		"host":          fmt.Sprintf("http://gimlet.%s:%s", os.Getenv("HOST"), "9000"),
		"adminToken":    gimletdAdminToken,
	}
}

func setGithubStackConfig(data *data) {
	data.github = true
	gimletDashboardConfig := data.stackConfig.Config["gimlet"].(map[string]interface{})

	gimletDashboardConfig["githubOrg"] = data.appOwner
	gimletDashboardConfig["githubAppId"] = data.id
	gimletDashboardConfig["githubPrivateKey"] = data.pem
	gimletDashboardConfig["githubClientId"] = data.clientId
	gimletDashboardConfig["githubClientSecret"] = data.clientSecret
	gimletDashboardConfig["githubInstallationId"] = data.installationId
}

func setGitlabStackConfig(data *data, token string) {
	data.gitlab = true
	gimletDashboardConfig := data.stackConfig.Config["gimlet"].(map[string]interface{})

	gimletDashboardConfig["gitlabOrg"] = data.appOwner
	gimletDashboardConfig["gitlabClientId"] = data.clientId
	gimletDashboardConfig["gitlabClientSecret"] = data.clientSecret
	gimletDashboardConfig["gitlabAdminToken"] = token
	gimletDashboardConfig["gitlabUrl"] = data.scmURL

	scmHost := strings.Split(data.scmURL, "://")[1]
	gimletDashboardConfig["gitSSHAddressFormat"] = "git@" + scmHost + ":%s.git"
}

func getContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := ctx.Value("data").(*data)

	_, err := token.ParseRequest(r, func(t *token.Token) (string, error) {
		return data.securityToken, nil
	})
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	context := map[string]interface{}{
		"appId":                   data.id,
		"clientId":                data.clientId,
		"clientSecret":            data.clientSecret,
		"pem":                     data.pem,
		"org":                     data.appOwner,
		"infraGitopsRepoFileName": data.infraGitopsRepoFileName,
		"infraPublicKey":          data.infraPublicKey,
		"infraSecretFileName":     data.infraSecretFileName,
		"appsGitopsRepoFileName":  data.appsGitopsRepoFileName,
		"appsPublicKey":           data.appsPublicKey,
		"appsSecretFileName":      data.appsSecretFileName,
		"notificationsFileName":   data.notificationsFileName,
		"infraRepo":               data.infraRepo,
		"appsRepo":                data.appsRepo,
		"repoPerEnv":              data.repoPerEnv,
		"envName":                 data.envName,
		"stackDefinition":         data.stackDefinition,
		"stackConfig":             data.stackConfig,
		"scmUrl":                  data.scmURL,
	}

	contextString, err := json.Marshal(context)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(contextString)
}

func created(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	url := "https://api.github.com/app-manifests/" + r.Form["code"][0] + "/conversions"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/vnd.github.fury-preview+json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}

	appInfo := map[string]interface{}{}
	err = json.Unmarshal(f, &appInfo)
	if err != nil {
		panic(err)
	}

	ctx := r.Context()
	data := ctx.Value("data").(*data)

	data.id = fmt.Sprintf("%.0f", appInfo["id"].(float64))
	data.clientId = appInfo["client_id"].(string)
	data.clientSecret = appInfo["client_secret"].(string)
	data.pem = appInfo["pem"].(string)
	data.slug = appInfo["slug"].(string)

	http.Redirect(w, r, fmt.Sprintf("https://github.com/apps/%s/installations/new", data.slug), http.StatusSeeOther)
}

func installed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	fmt.Println(formValues)

	ctx := r.Context()
	data := ctx.Value("data").(*data)
	data.installationId = formValues.Get("installation_id")

	privateKey := config.Multiline(data.pem)
	tokenManager, err := customGithub.NewGithubOrgTokenManager(data.id, data.installationId, privateKey.String())
	if err != nil {
		panic(err)
	}
	tokenString, err := tokenManager.AppToken()
	if err != nil {
		panic(err)
	}

	gitSvc := &customGithub.GithubClient{}
	appOwner, err := gitSvc.GetAppOwner(tokenString)
	if err != nil {
		logrus.Errorf("cannot get app info: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data.appOwner = appOwner
	data.tokenManager = tokenManager

	http.Redirect(w, r, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s", data.clientId), http.StatusSeeOther)
}

func auth(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	ctx := r.Context()
	data := ctx.Value("data").(*data)
	url := fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		data.clientId,
		data.clientSecret,
		r.Form["code"][0],
	)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/vnd.github.fury-preview+json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}

	appInfo := map[string]interface{}{}
	err = json.Unmarshal(f, &appInfo)
	if err != nil {
		panic(err)
	}

	data.accessToken = appInfo["access_token"].(string)
	data.refreshToken = appInfo["refresh_token"].(string)

	goScmHelper := genericScm.NewGoScmHelper(&config.Config{
		Github: config.Github{AppID: "dummyID - helper uses the tokens only"},
	}, nil)
	scmUser, err := goScmHelper.User(data.accessToken, data.refreshToken)
	if err != nil {
		panic(err)
	}
	data.loggedInUser = scmUser.Login

	initStackConfig(data)

	data.scmURL = "https://github.com"
	setGithubStackConfig(data)

	data.securityToken = uuid.New().String()
	setSessionCookie(w, r, data.securityToken)

	http.Redirect(w, r, "/step-2", http.StatusSeeOther)
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, securityToken string) error {
	sixHours, _ := time.ParseDuration("6h")
	exp := time.Now().Add(sixHours).Unix()
	t := token.New(token.SessToken, securityToken)
	tokenStr, err := t.SignExpires(securityToken, exp)
	if err != nil {
		return err
	}

	httputil.SetCookie(w, r, "user_sess", tokenStr)
	return nil
}

func bootstrap(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	formValues := r.PostForm

	ctx := r.Context()
	data := ctx.Value("data").(*data)

	installationToken, gitUser, err := data.tokenManager.Token()
	if err != nil {
		panic(err)
	}

	var gitSvc customScm.CustomGitService
	if data.github {
		gitSvc = &customGithub.GithubClient{}
	} else {
		gitSvc = &customGitlab.GitlabClient{
			BaseURL: data.scmURL,
		}
	}

	repos, err := gitSvc.OrgRepos(installationToken)
	if err != nil {
		panic(err)
	}

	infraRepo := formValues.Get("infra")
	appsRepo := formValues.Get("apps")
	envName := formValues.Get("env")
	repoPerEnv, err := strconv.ParseBool(formValues.Get("repoPerEnv"))
	if err != nil {
		panic(err)
	}

	if !strings.Contains(infraRepo, "/") {
		infraRepo = filepath.Join(data.appOwner, infraRepo)
	}
	if !strings.Contains(appsRepo, "/") {
		appsRepo = filepath.Join(data.appOwner, appsRepo)
	}

	data.infraRepo = infraRepo
	data.appsRepo = appsRepo
	data.repoPerEnv = repoPerEnv
	data.envName = envName

	_, err = server.AssureRepoExists(
		repos,
		infraRepo,
		data.accessToken,
		installationToken,
		data.loggedInUser,
		gitSvc,
	)
	if err != nil {
		panic(err)
	}
	_, err = server.AssureRepoExists(
		repos,
		appsRepo,
		data.accessToken,
		installationToken,
		data.loggedInUser,
		gitSvc,
	)
	if err != nil {
		panic(err)
	}

	repoCachePath, err := ioutil.TempDir("", "cache")
	if err != nil {
		panic(err)
	}

	fakeDashConfig := &config.Config{}
	if data.github {
		fakeDashConfig.Github = config.Github{AppID: "xxx"}
	} else {
		fakeDashConfig.Gitlab = config.Gitlab{
			ClientID: "xxx",
			URL:      data.scmURL,
		}
	}
	gitRepoCache, err := nativeGit.NewRepoCache(
		data.tokenManager,
		nil,
		repoCachePath,
		nil,
		fakeDashConfig,
		nil,
	)
	if err != nil {
		panic(err)
	}
	go gitRepoCache.Run()
	data.repoCache = gitRepoCache

	infraGitopsRepoFileName, infraSecretFileName, err := server.BootstrapEnv(
		gitRepoCache,
		gitSvc,
		envName,
		infraRepo,
		repoPerEnv,
		installationToken,
		true,
		true,
		false,
		false,
		data.scmURL,
	)
	if err != nil {
		panic(err)
	}
	appsGitopsRepoFileName, appsSecretFileName, err := server.BootstrapEnv(
		gitRepoCache,
		gitSvc,
		envName,
		appsRepo,
		repoPerEnv,
		installationToken,
		false,
		false,
		true,
		true,
		data.scmURL,
	)
	if err != nil {
		panic(err)
	}

	gimletDashboardConfig := data.stackConfig.Config["gimlet"].(map[string]interface{})
	gimletURL := "http://gimlet.infrastructure.svc.cluster.local:8888"
	gimletSignedAdminToken := gimletDashboardConfig["gimletdToken"].(string)

	notificationsFileName, err := server.BootstrapNotifications(
		gitRepoCache,
		gimletURL,
		gimletSignedAdminToken,
		envName,
		appsRepo,
		repoPerEnv,
		installationToken,
		gitUser,
		data.scmURL,
	)
	if err != nil {
		panic(err)
	}

	gimletDashboardConfig["bootstrapEnv"] = fmt.Sprintf("name=%s&repoPerEnv=%t&infraRepo=%s&appsRepo=%s", envName, repoPerEnv, infraRepo, appsRepo)

	gimletAgentConfig := data.stackConfig.Config["gimletAgent"].(map[string]interface{})
	gimletAgentConfig["environment"] = envName

	data.infraGitopsRepoFileName = infraGitopsRepoFileName
	data.infraSecretFileName = infraSecretFileName

	data.appsGitopsRepoFileName = appsGitopsRepoFileName
	data.appsSecretFileName = appsSecretFileName

	data.notificationsFileName = notificationsFileName

	stackConfigBuff := bytes.NewBufferString("")
	e := yaml.NewEncoder(stackConfigBuff)
	e.SetIndent(2)
	err = e.Encode(data.stackConfig)
	if err != nil {
		panic(err)
	}

	stackYamlPath := filepath.Join(envName, "stack.yaml")
	if repoPerEnv {
		stackYamlPath = "stack.yaml"
	}

	repo, tmpPath, err := gitRepoCache.InstanceForWrite(infraRepo)
	defer os.RemoveAll(tmpPath)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), nativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = stack.GenerateAndWriteFiles(*data.stackConfig, filepath.Join(tmpPath, stackYamlPath))
	if err != nil {
		logrus.Errorf("cannot generate and write files: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = server.StageCommitAndPush(repo, tmpPath, installationToken, "[Gimlet Dashboard] Updating components")
	if err != nil {
		logrus.Errorf("cannot stage commit and push: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/step-3", http.StatusSeeOther)
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func writeTempFiles(workDir string) {
	ioutil.WriteFile(filepath.Join(workDir, "index.html"), web.IndexHtml, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "main.js"), web.MainJs, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "main.css"), web.MainCSS, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "1.chunk.js"), web.ChunkJs, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "favicon.ico"), web.Favicon, 0666)
}

func removeTempFiles(workDir string) {
	os.Remove(workDir)
}

func done(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

func loadStackDefinition(stackConfig *dx.StackConfig) (map[string]interface{}, error) {
	stackDefinitionYaml, err := stack.StackDefinitionFromRepo(stackConfig.Stack.Repository)
	if err != nil {
		return nil, fmt.Errorf("cannot get stack definition: %s", err.Error())
	}

	var stackDefinition map[string]interface{}
	err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
	return stackDefinition, err
}

func gitlabInit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	fmt.Println(formValues)

	gitlabUrl := formValues.Get("gitlabUrl")
	token := formValues.Get("token")
	appId := formValues.Get("appId")
	appSecret := formValues.Get("appSecret")

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabUrl))
	if err != nil {
		panic(err)
	}

	user, _, err := git.Users.CurrentUser()
	if err != nil {
		panic(err)
	}

	var org string
	if user.Bot {
		groups, _, err := git.Groups.ListGroups(&gitlab.ListGroupsOptions{})
		if err != nil {
			panic(err)
		}
		org = groups[0].FullPath
	} else {
		org = user.Username
	}

	ctx := r.Context()
	data := ctx.Value("data").(*data)

	data.id = appId
	data.clientId = appId
	data.clientSecret = appSecret

	tokenManager := customGitlab.NewGitlabTokenManager(token)

	data.appOwner = org
	data.tokenManager = tokenManager

	data.loggedInUser = user.Username

	initStackConfig(data)

	data.scmURL = gitlabUrl
	setGitlabStackConfig(data, token)

	data.securityToken = uuid.New().String()
	setSessionCookie(w, r, data.securityToken)

	http.Redirect(w, r, "/step-2", http.StatusSeeOther)
}

func groupFromUsername(username string) string {
	username = strings.TrimPrefix(username, "group_")
	return strings.Split(username, "_")[0]
}
