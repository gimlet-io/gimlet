package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

	// GITHUB_APP_ID
	id := fmt.Sprintf("%.0f", appInfo["id"].(float64))
	// GITHUB_CLIENT_ID
	clientId := appInfo["client_id"].(string)
	// GITHUB_CLIENT_SECRET
	clientSecret := appInfo["client_secret"].(string)
	// GITHUB_PRIVATE_KEY
	pem := appInfo["pem"].(string)
	fmt.Println(id)
	fmt.Println(clientId)
	fmt.Println(clientSecret)
	fmt.Println(pem)
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

	// GITHUB_INSTALLATION_ID
	installationId := formValues.Get("installation_id")
	fmt.Println(installationId)

	// // TODO pem, id, installationId will come from the persistentConfig
	// privateKey := config.Multiline(data.pem)
	// tokenManager, err := customGithub.NewGithubOrgTokenManager(data.id, data.installationId, privateKey.String())
	// if err != nil {
	// 	panic(err)
	// }
	// tokenString, err := tokenManager.AppToken()
	// if err != nil {
	// 	panic(err)
	// }

	// gitSvc := &customGithub.GithubClient{}
	// // GITHUB_ORG
	// appOwner, err := gitSvc.GetAppOwner(tokenString)
	// if err != nil {
	// 	logrus.Errorf("cannot get app info: %s", err)
	// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	// 	return
	// }

	http.Redirect(w, r, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s", "TODO config clientId"), http.StatusSeeOther)
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

	// TODO save the config to persistentConfig
	// GITLAB_CLIENT_ID
	fmt.Println(appId)
	// GITLAB_CLIENT_SECRET
	fmt.Println(appSecret)
	// GITLAB_ORG
	fmt.Println(org)
	// GITLAB_URL
	fmt.Println(gitlabUrl)
	// GITLAB_ADMIN_TOKEN
	fmt.Println(token)

	http.Redirect(w, r, "/repositories", http.StatusSeeOther)
}
