package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

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

	// TODO error handling
	ctx := r.Context()
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)
	config.Save(store.GithubAppID, fmt.Sprintf("%.0f", appInfo["id"].(float64)))
	config.Save(store.GithubClientID, appInfo["client_id"].(string))
	config.Save(store.GithubClientSecret, appInfo["client_secret"].(string))
	config.Save(store.GithubPrivateKey, appInfo["pem"].(string))
	slug := appInfo["slug"].(string)

	http.Redirect(w, r, fmt.Sprintf("https://github.com/apps/%s/installations/new", slug), http.StatusSeeOther)
}

func installed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	fmt.Println(formValues)
	installationId := formValues.Get("installation_id")

	ctx := r.Context()
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)
	// TODO error handling
	config.Save(store.GithubInstallationID, installationId)

	tokenManager, err := customGithub.NewGithubOrgTokenManager(config.Get(store.GithubAppID), installationId, config.Get(store.GithubPrivateKey))
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
	// TODO error handling
	config.Save(store.GithubOrg, appOwner)

	gitServiceImplFromCtx := ctx.Value("gitService").(*customScm.CustomGitService)
	tokenManagerFromCtx := ctx.Value("tokenManager").(*customScm.NonImpersonatedTokenManager)
	*gitServiceImplFromCtx = &customGithub.GithubClient{}
	*tokenManagerFromCtx = tokenManager
	// TODO admin user need access to repos

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
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

	// TODO error handling
	ctx := r.Context()
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)
	config.Save(store.GitlabClientID, appId)
	config.Save(store.GitlabClientSecret, appSecret)
	config.Save(store.GitlabOrg, org)
	config.Save(store.GitlabURL, gitlabUrl)
	config.Save(store.GitlabAdminToken, token)

	gitServiceImplFromCtx := ctx.Value("gitService").(*customScm.CustomGitService)
	tokenManagerFromCtx := ctx.Value("tokenManager").(*customScm.NonImpersonatedTokenManager)
	(*gitServiceImplFromCtx) = &customGitlab.GitlabClient{
		BaseURL: gitlabUrl,
	}
	(*tokenManagerFromCtx) = customGitlab.NewGitlabTokenManager(token)

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
