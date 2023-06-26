package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/go-chi/chi"
	"github.com/laszlocph/go-login/login"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	dynamicConfig.Github.AppID = fmt.Sprintf("%.0f", appInfo["id"].(float64))
	dynamicConfig.Github.ClientID = appInfo["client_id"].(string)
	dynamicConfig.Github.ClientSecret = appInfo["client_secret"].(string)
	dynamicConfig.Github.PrivateKey.Decode(appInfo["pem"].(string))
	dynamicConfig.Persist()
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
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	dynamicConfig.Github.InstallationID = installationId

	token, err := exchange(
		formValues["code"][0],
		"",
		dynamicConfig.Github.ClientID,
		dynamicConfig.Github.ClientSecret,
		"",
	)
	if err != nil {
		panic(err)
	}

	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, nil)
	scmUser, err := goScmHelper.User(token.Access, token.Refresh)
	if err != nil {
		log.Errorf("cannot find git user: %s", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	dao := ctx.Value("store").(*store.Store)
	user, err := getOrCreateUser(dao, scmUser, token)
	if err != nil {
		log.Errorf("cannot get or store user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = setSessionCookie(w, r, user)
	if err != nil {
		log.Errorf("cannot set session cookie: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenManager, err := customGithub.NewGithubOrgTokenManager(dynamicConfig.Github.AppID, installationId, dynamicConfig.Github.PrivateKey.String())
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

	dynamicConfig.Github.Org = appOwner
	dynamicConfig.Persist()

	gitScm := ctx.Value("gitScm").(*customScm.GitScm)
	gitScm.GitService = gitSvc
	gitScm.TokenManager = tokenManager

	goScm := ctx.Value("goScm").(*genericScm.GoScmHelper)
	*goScm = *genericScm.NewGoScmHelper(dynamicConfig, nil)

	router := ctx.Value("router").(*chi.Mux)
	config := ctx.Value("config").(*config.Config)
	githubOAuthRoutes(dynamicConfig, router, config.Host)

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

	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	dynamicConfig.Gitlab.ClientID = appId
	dynamicConfig.Gitlab.ClientSecret = appSecret
	dynamicConfig.Gitlab.Org = org
	dynamicConfig.Gitlab.URL = gitlabUrl
	dynamicConfig.Gitlab.AdminToken = token
	dynamicConfig.Persist()

	gitScm := ctx.Value("gitScm").(*customScm.GitScm)
	gitScm.GitService = &customGitlab.GitlabClient{
		BaseURL: gitlabUrl,
	}
	gitScm.TokenManager = customGitlab.NewGitlabTokenManager(token)

	goScm := ctx.Value("goScm").(*genericScm.GoScmHelper)
	*goScm = *genericScm.NewGoScmHelper(dynamicConfig, nil)

	router := ctx.Value("router").(*chi.Mux)
	config := ctx.Value("config").(*config.Config)
	githubOAuthRoutes(dynamicConfig, router, config.Host)

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// exchange converts an authorization code into a token.
func exchange(
	code,
	state,
	clientID,
	clientSecret,
	redirectURL string) (*login.Token, error) {
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	if len(state) != 0 {
		v.Set("state", state)
	}

	if len(redirectURL) != 0 {
		v.Set("redirect_uri", redirectURL)
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// req.SetBasicAuth(clientID, clientSecret)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		err := new(Error)
		json.NewDecoder(res.Body).Decode(err)
		return nil, err
	}

	token := &githubToken{}
	err = json.NewDecoder(res.Body).Decode(token)

	return &login.Token{
		Access:  token.AccessToken,
		Refresh: token.RefreshToken}, err
}

// Error represents a failed authorization request.
type Error struct {
	Code string `json:"error"`
	Desc string `json:"error_description"`
}

// Error returns the string representation of an
// authorization error.
func (e *Error) Error() string {
	return e.Code + ": " + e.Desc
}

// token stores the authorization credentials used to
// access protected resources.
type githubToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expires      int64  `json:"expires_in"`
}
