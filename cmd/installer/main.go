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

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/installer/web"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"encoding/hex"
	"math/rand"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.WithValue("data", &data{
		org: os.Getenv("ORG"),
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

	srv := http.Server{Addr: ":443", Handler: r}
	go func() {
		// err = srv.ListenAndServe()
		err = srv.ListenAndServeTLS(filepath.Join(workDir, "server.crt"), filepath.Join(workDir, "server.key"))
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

type data struct {
	id                      string
	slug                    string
	clientId                string
	clientSecret            string
	pem                     string
	org                     string
	installationId          string
	tokenManager            *customGithub.GithubOrgTokenManager
	accessToken             string
	refreshToken            string
	loggedInUser            string
	repoCache               *nativeGit.RepoCache
	gimletdPublicKey        string
	isNewInfraRepo          bool
	isNewAppsRepo           bool
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
}

func getContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := ctx.Value("data").(*data)

	context := map[string]interface{}{
		"appId":                   data.id,
		"clientId":                data.clientId,
		"clientSecret":            data.clientSecret,
		"pem":                     data.pem,
		"org":                     data.org,
		"gimletdPublicKey":        data.gimletdPublicKey,
		"isNewInfraRepo":          data.isNewInfraRepo,
		"isNewAppsRepo":           data.isNewAppsRepo,
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

	tokenManager, err := customGithub.NewGithubOrgTokenManager(&config.Config{
		Github: config.Github{
			AppID:          data.id,
			PrivateKey:     config.Multiline(data.pem),
			InstallationID: data.installationId,
		},
	})
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
	data.org = appOwner
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
		Github: config.Github{
			ClientID:     data.clientId,
			ClientSecret: data.clientSecret,
		},
	}, nil)
	scmUser, err := goScmHelper.User(data.accessToken, data.refreshToken)
	if err != nil {
		panic(err)
	}
	data.loggedInUser = scmUser.Login

	http.Redirect(w, r, "/step-2", http.StatusSeeOther)
}

func bootstrap(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	formValues := r.PostForm
	fmt.Println(formValues)

	ctx := r.Context()
	data := ctx.Value("data").(*data)

	installationToken, gitUser, err := data.tokenManager.Token()
	if err != nil {
		panic(err)
	}

	gitSvc := &customGithub.GithubClient{}
	repos, err := gitSvc.OrgRepos(installationToken)
	if err != nil {
		panic(err)
	}

	infraRepo := formValues.Get("infra")
	appsRepo := formValues.Get("apps")
	envName := formValues.Get("env")
	email := formValues.Get("email")
	repoPerEnv, err := strconv.ParseBool(formValues.Get("repoPerEnv"))
	if err != nil {
		panic(err)
	}

	if !strings.Contains(infraRepo, "/") {
		infraRepo = filepath.Join(data.org, infraRepo)
	}
	if !strings.Contains(appsRepo, "/") {
		appsRepo = filepath.Join(data.org, appsRepo)
	}

	data.infraRepo = infraRepo
	data.appsRepo = appsRepo
	data.repoPerEnv = repoPerEnv
	data.envName = envName

	isNewInfraRepo, err := server.AssureRepoExists(
		repos,
		infraRepo,
		data.accessToken,
		installationToken,
		data.loggedInUser,
	)
	if err != nil {
		panic(err)
	}
	isNewAppsRepo, err := server.AssureRepoExists(
		repos,
		appsRepo,
		data.accessToken,
		installationToken,
		data.loggedInUser,
	)
	if err != nil {
		panic(err)
	}
	data.isNewInfraRepo = isNewInfraRepo
	data.isNewAppsRepo = isNewAppsRepo

	repoCachePath, err := ioutil.TempDir("", "cache")
	if err != nil {
		panic(err)
	}
	gitRepoCache, err := nativeGit.NewRepoCache(
		data.tokenManager,
		nil,
		repoCachePath,
		nil,
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}
	go gitRepoCache.Run()
	data.repoCache = gitRepoCache

	infraGitopsRepoFileName, infraPublicKey, infraSecretFileName, err := server.BootstrapEnv(
		gitRepoCache,
		envName,
		infraRepo,
		repoPerEnv,
		installationToken,
		true,
	)
	if err != nil {
		panic(err)
	}
	appsGitopsRepoFileName, appsPublicKey, appsSecretFileName, err := server.BootstrapEnv(
		gitRepoCache,
		envName,
		appsRepo,
		repoPerEnv,
		installationToken,
		false,
	)
	if err != nil {
		panic(err)
	}

	gimletdUrl := "https://gimletd." + os.Getenv("HOST")

	gimletdAdminToken := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	token := token.New(token.UserToken, "admin")
	gimletdSignedAdminToken, err := token.Sign(gimletdAdminToken)
	if err != nil {
		panic(err)
	}

	notificationsFileName, err := server.BootstrapNotifications(
		gitRepoCache,
		gimletdUrl,
		gimletdSignedAdminToken,
		envName,
		appsRepo,
		repoPerEnv,
		installationToken,
		gitUser,
	)
	if err != nil {
		panic(err)
	}

	data.infraGitopsRepoFileName = infraGitopsRepoFileName
	data.infraPublicKey = infraPublicKey
	data.infraSecretFileName = infraSecretFileName

	data.appsGitopsRepoFileName = appsGitopsRepoFileName
	data.appsPublicKey = appsPublicKey
	data.appsSecretFileName = appsSecretFileName

	data.notificationsFileName = notificationsFileName

	jwtSecret, _ := randomHex(32)
	agentAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)
	_, agentToken, _ := agentAuth.Encode(map[string]interface{}{"user_id": "gimlet-agent"})

	webhookSecret, _ := randomHex(32)
	privateKeyBytes, publicKeyBytes, err := gitops.GenerateEd25519()
	if err != nil {
		panic(err)
	}
	data.gimletdPublicKey = string(publicKeyBytes)

	bootstrapEnv := fmt.Sprintf(
		"name=%s&repoPerEnv=%t&infraRepo=%s&appsRepo=%s",
		data.envName,
		data.repoPerEnv,
		data.infraRepo,
		data.appsRepo,
	)

	useExistingPostgres, err := strconv.ParseBool(formValues.Get("useExistingPostgres"))
	if err != nil {
		panic(err)
	}

	var dashboardPostgresConfig map[string]interface{}
	var gimletdPostgresConfig map[string]interface{}
	if useExistingPostgres {
		dashboardPostgresConfig = map[string]interface{}{
			"hostAndPort": formValues.Get("hostAndPort"),
			"db":          formValues.Get("dashboardDb"),
			"user":        formValues.Get("dashboardUsername"),
			"password":    formValues.Get("dashboardPassword"),
		}
		gimletdPostgresConfig = map[string]interface{}{
			"hostAndPort": formValues.Get("hostAndPort"),
			"db":          formValues.Get("gimletdDb"),
			"user":        formValues.Get("gimletdUsername"),
			"password":    formValues.Get("gimletdPassword"),
		}
	} else {
		postgresPassword := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
		dashboardPassword := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
		gimletdPassword := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

		dashboardPostgresConfig = map[string]interface{}{
			"install":          true,
			"hostAndPort":      "postgresql:5432",
			"postgresPassword": postgresPassword,
			"db":               "gimlet_dashboard",
			"user":             "gimlet_dashboard",
			"password":         dashboardPassword,
		}
		gimletdPostgresConfig = map[string]interface{}{
			"install":          true,
			"hostAndPort":      "postgresql:5432",
			"postgresPassword": postgresPassword,
			"db":               "gimletd",
			"user":             "gimletd",
			"password":         gimletdPassword,
		}
	}

	stackConfig := &dx.StackConfig{
		Stack: dx.StackRef{
			Repository: stack.DefaultStackURL,
		},
		Config: map[string]interface{}{
			"civo": map[string]interface{}{
				"enabled": true,
			},
			"nginx": map[string]interface{}{
				"enabled": true,
				"host":    os.Getenv("HOST"),
			},
			"certManager": map[string]interface{}{
				"enabled": true,
				"email":   email,
			},
			"gimletd": map[string]interface{}{
				"enabled":    true,
				"gitopsRepo": appsRepo,
				"deployKey":  string(privateKeyBytes),
				"adminToken": gimletdAdminToken,
				"postgresql": gimletdPostgresConfig,
			},
			"gimletAgent": map[string]interface{}{
				"enabled":          true,
				"environment":      envName,
				"dashboardAddress": "https://gimlet." + os.Getenv("HOST"),
				"agentKey":         agentToken,
			},
			"gimletDashboard": map[string]interface{}{
				"enabled":              true,
				"jwtSecret":            jwtSecret,
				"githubOrg":            data.org,
				"gimletdToken":         gimletdSignedAdminToken,
				"githubAppId":          data.id,
				"githubPrivateKey":     data.pem,
				"githubClientId":       data.clientId,
				"githubClientSecret":   data.clientSecret,
				"webhookSecret":        webhookSecret,
				"githubInstallationId": data.installationId,
				"bootstrapEnv":         bootstrapEnv,
				"postgresql":           dashboardPostgresConfig,
			},
		},
	}

	latestTag, _ := stack.LatestVersion(stackConfig.Stack.Repository)
	if latestTag != "" {
		stackConfig.Stack.Repository = stackConfig.Stack.Repository + "?tag=" + latestTag
	}

	stackConfigBuff := bytes.NewBufferString("")
	e := yaml.NewEncoder(stackConfigBuff)
	e.SetIndent(2)
	err = e.Encode(stackConfig)
	if err != nil {
		panic(err)
	}

	stackYamlPath := filepath.Join(envName, "stack.yaml")
	if repoPerEnv {
		stackYamlPath = "stack.yaml"
	}

	repo, tmpPath, err := gitRepoCache.InstanceForWrite(infraRepo)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpPath)

	err = os.WriteFile(filepath.Join(tmpPath, stackYamlPath), stackConfigBuff.Bytes(), nativeGit.Dir_RWX_RX_R)
	if err != nil {
		logrus.Errorf("cannot write file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = stack.GenerateAndWriteFiles(*stackConfig, filepath.Join(tmpPath, stackYamlPath))
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
	ioutil.WriteFile(filepath.Join(workDir, "server.crt"), web.ServerCrt, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "server.key"), web.ServerKey, 0666)
}

func removeTempFiles(workDir string) {
	os.Remove(workDir)
}

func done(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}
