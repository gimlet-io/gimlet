package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dash_config "github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
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

	ctx := r.Context()
	persistentConfig := ctx.Value("config").(*dash_config.PersistentConfig)
	persistentConfig.Save(&dash_config.Config{
		Github: dash_config.Github{
			AppID:        appInfo["id"].(string),
			PrivateKey:   appInfo["pem"].(dash_config.Multiline),
			ClientID:     appInfo["client_id"].(string),
			ClientSecret: appInfo["client_secret"].(string),
		},
	})

	http.Redirect(w, r, fmt.Sprintf("https://github.com/apps/%s/installations/new", appInfo["slug"].(string)), http.StatusSeeOther)
}

func installed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	fmt.Println(formValues)

	ctx := r.Context()
	persistentConfig := ctx.Value("config").(*dash_config.PersistentConfig)
	config, err := persistentConfig.Get()
	if err != nil {
		logrus.Errorf("cannot get persistent config: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenManager, err := customGithub.NewGithubOrgTokenManager(config.Github.AppID, config.Github.InstallationID, config.Github.PrivateKey.String())
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

	persistentConfig.Save(&dash_config.Config{
		Github: dash_config.Github{
			InstallationID: formValues.Get("installation_id"),
			Org:            appOwner,
		},
	})

	http.Redirect(w, r, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s", config.Github.ClientID), http.StatusSeeOther)
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
	persistentConfig := ctx.Value("config").(*dash_config.PersistentConfig)
	persistentConfig.Save(&dash_config.Config{
		Gitlab: dash_config.Gitlab{
			ClientID:     appId,
			ClientSecret: appSecret,
			AdminToken:   token,
			Org:          org,
			URL:          gitlabUrl,
		},
	})

	http.Redirect(w, r, "/repositories", http.StatusSeeOther)
}
