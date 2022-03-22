package main

import (
	"bytes"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fluxcd/pkg/ssh"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"crypto/rand"
	"encoding/hex"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.WithValue("data", &data{}))

	r.Get("/created", created)
	r.Get("/auth", auth)
	r.Get("/installed", installed)
	r.Post("/bootstrap", bootstrap)
	r.HandleFunc("/*", serveTemplate)

	// err := http.ListenAndServe(":4443", r)
	err := http.ListenAndServeTLS(":4443", "server.crt", "server.key", r)
	fmt.Println(err)
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
	infraRepo               string
	appsRepo                string
	repoPerEnv              bool
	envName                 string
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "index.html"
	} else if !strings.HasSuffix(path, ".html") {
		path = path + ".html"
	}
	fp := filepath.Join("web", path)

	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	ctx := r.Context()
	data := ctx.Value("data").(*data)

	err = tmpl.Execute(w, map[string]interface{}{
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
		"infraRepo":               data.infraRepo,
		"appsRepo":                data.appsRepo,
		"repoPerEnv":              data.repoPerEnv,
		"envName":                 data.envName,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
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

	tokenString, _, err := data.tokenManager.Token()
	if err != nil {
		panic(err)
	}

	gitSvc := &customGithub.GithubClient{}
	repos, err := gitSvc.OrgRepos(tokenString)
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
		infraRepo = filepath.Join(data.org, infraRepo)
	}
	if !strings.Contains(appsRepo, "/") {
		appsRepo = filepath.Join(data.org, appsRepo)
	}

	data.infraRepo = infraRepo
	data.appsRepo = appsRepo
	data.repoPerEnv = repoPerEnv
	data.envName = envName

	isNewInfraRepo, err := server.AssureRepoExists(repos, infraRepo, data.accessToken)
	if err != nil {
		panic(err)
	}
	isNewAppsRepo, err := server.AssureRepoExists(repos, appsRepo, data.accessToken)
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

	infraGitopsRepoFileName, infraPublicKey, infraSecretFileName, err := server.BootstrapEnv(gitRepoCache, envName, infraRepo, repoPerEnv, tokenString)
	if err != nil {
		panic(err)
	}
	appsGitopsRepoFileName, appsPublicKey, appsSecretFileName, err := server.BootstrapEnv(gitRepoCache, envName, appsRepo, repoPerEnv, tokenString)
	if err != nil {
		panic(err)
	}

	data.infraGitopsRepoFileName = infraGitopsRepoFileName
	data.infraPublicKey = infraPublicKey
	data.infraSecretFileName = infraSecretFileName

	data.appsGitopsRepoFileName = appsGitopsRepoFileName
	data.appsPublicKey = appsPublicKey
	data.appsSecretFileName = appsSecretFileName

	jwtSecret, _ := randomHex(32)
	agentAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)
	_, agentToken, _ := agentAuth.Encode(map[string]interface{}{"user_id": "gimlet-agent"})

	webhookSecret, _ := randomHex(32)
	keyPair, err := ssh.NewEd25519Generator().Generate()
	if err != nil {
		panic(err)
	}
	data.gimletdPublicKey = string(keyPair.PublicKey)

	gimletdAdminToken := base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

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
			"hostAndPort":      "postgres:5432",
			"postgresPassword": postgresPassword,
			"db":               "gimlet_dashboard",
			"user":             "gimlet_dashboard",
			"password":         dashboardPassword,
		}
		gimletdPostgresConfig = map[string]interface{}{
			"install":          true,
			"hostAndPort":      "postgres:5432",
			"postgresPassword": postgresPassword,
			"db":               "gimletd",
			"user":             "gimletd",
			"password":         gimletdPassword,
		}
	}

	stackConfig := &dx.StackConfig{
		Stack: dx.StackRef{
			Repository: "https://github.com/gimlet-io/gimlet-stack-reference.git?branch=gimlet-in-stack",
		},
		Config: map[string]interface{}{
			"nginx": map[string]interface{}{
				"enabled": true,
				"host":    os.Getenv("HOST"),
			},
			"gimletd": map[string]interface{}{
				"enabled":    true,
				"gitopsRepo": appsRepo,
				"deployKey":  string(keyPair.PrivateKey),
				"adminToken": gimletdAdminToken,
				"postgresql": gimletdPostgresConfig,
			},
			"gimletAgent": map[string]interface{}{
				"enabled":     true,
				"environment": envName,
				"agentKey":    agentToken,
			},
			"gimletDashboard": map[string]interface{}{
				"enabled":              true,
				"jwtSecret":            jwtSecret,
				"githubOrg":            data.org,
				"gimletdToken":         gimletdAdminToken,
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

	err = server.StageCommitAndPush(repo, tokenString, "[Gimlet Dashboard] Updating components")
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
