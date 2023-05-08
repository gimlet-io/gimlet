package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	//TODO save to DB through PersistentConfig

	// data.id = fmt.Sprintf("%.0f", appInfo["id"].(float64))
	// data.clientId = appInfo["client_id"].(string)
	// data.clientSecret = appInfo["client_secret"].(string)
	// data.pem = appInfo["pem"].(string)
	// data.slug = appInfo["slug"].(string)

	http.Redirect(w, r, fmt.Sprintf("https://github.com/apps/%s/installations/new", appInfo["slug"].(string)), http.StatusSeeOther)
}

func installed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	fmt.Println(formValues)

	// TODO save installationID to DB through PersistentConfig

	// data.installationId = formValues.Get("installation_id")

	// TODO get the org or app owner

	// tokenString, err := tokenManager.AppToken()
	// if err != nil {
	// 	panic(err)
	// }
	// gitSvc := &customGithub.GithubClient{}
	// appOwner, err := gitSvc.GetAppOwner(tokenString)
	// if err != nil {
	// 	logrus.Errorf("cannot get app info: %s", err)
	// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	// 	return
	// }
	// data.appOwner = appOwner

	// TODO get the client id
	http.Redirect(w, r, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s", "TODO clientId"), http.StatusSeeOther)
}
