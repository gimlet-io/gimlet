package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
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

	// http.ListenAndServe(":3333", r)
	err := http.ListenAndServeTLS(":443", "server.crt", "server.key", r)
	fmt.Println(err)
}

type data struct {
	id               string
	slug             string
	clientId         string
	clientSecret     string
	pem              string
	org              string
	installationId   string
	tokenManager     *customGithub.GithubOrgTokenManager
	accessToken      string
	refreshToken     string
	repoCache        *nativeGit.RepoCache
	gimletdPublicKey string
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
		"appId":        data.id,
		"clientId":     data.clientId,
		"clientSecret": data.clientSecret,
		"pem":          data.pem,
		"org":          data.org,
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

	_, err = server.AssureRepoExists(repos, infraRepo, data.accessToken)
	if err != nil {
		panic(err)
	}
	_, err = server.AssureRepoExists(repos, appsRepo, data.accessToken)
	if err != nil {
		panic(err)
	}

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

	_, _, _, err = server.BootstrapEnv(gitRepoCache, envName, infraRepo, repoPerEnv, tokenString)
	if err != nil {
		panic(err)
	}
	_, _, _, err = server.BootstrapEnv(gitRepoCache, envName, appsRepo, repoPerEnv, tokenString)
	if err != nil {
		panic(err)
	}

	jwtSecret, _ := randomHex(32)
	agentAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)
	_, agentToken, _ := agentAuth.Encode(map[string]interface{}{"user_id": "gimlet-agent"})

	webhookSecret, _ := randomHex(32)
	privateKeyBytes, publicKeyBytes := gitops.GenerateKeyPair()
	data.gimletdPublicKey = string(publicKeyBytes)

	stackConfig := &dx.StackConfig{
		Stack: dx.StackRef{
			Repository: "https://github.com/gimlet-io/gimlet-stack-reference.git?branch=gimlet-in-stack",
		},
		Config: map[string]interface{}{
			"nginx": map[string]interface{}{
				"enabled": true,
				"host":    r.Host,
			},
			"gimletd": map[string]interface{}{
				"enabled":    true,
				"gitopsRepo": appsRepo,
				"deployKey":  string(privateKeyBytes),
			},
			"gimletAgent": map[string]interface{}{
				"enabled":     true,
				"environment": envName,
				"agentKey":    agentToken,
			},
			"gimletDashboard": map[string]interface{}{
				"enabled":            true,
				"jwtSecret":          jwtSecret,
				"githubOrg":          data.org,
				"gimletdToken":       "TODO",
				"githubAppId":        data.id,
				"githubPrivateKey":   data.pem,
				"githubClientId":     data.clientId,
				"githubClientSecret": data.clientSecret,
				"webhookSecret":      webhookSecret,
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
