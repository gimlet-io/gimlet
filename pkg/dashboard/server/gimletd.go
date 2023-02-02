package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func gitopsRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)

	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}

	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)

	client := client.NewClient(config.GimletD.URL, auth)
	gitopsRepo, err := client.GitopsRepoGet()
	if err != nil {
		logrus.Errorf("cannot get gitops repo: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	gitopsRepoString, err := json.Marshal(map[string]interface{}{
		"gitopsRepo": gitopsRepo,
	})
	if err != nil {
		logrus.Errorf("cannot serialize gitopsRepo: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(gitopsRepoString)
}

func gimletdBasicInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	config := ctx.Value("config").(*config.Config)

	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)

	client := client.NewClient(config.GimletD.URL, auth)
	gimletdUser, err := client.UserGet(user.Login, true)
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		gimletdUser, err = client.UserPost(&model.User{Login: user.Login})
	}
	if err != nil {
		logrus.Errorf("cannot get GimletD user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userString, err := json.Marshal(map[string]interface{}{
		"url":  config.GimletD.URL,
		"user": gimletdUser,
	})
	if err != nil {
		logrus.Errorf("cannot serialize user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userString)
}

func saveUser(w http.ResponseWriter, r *http.Request) {
	var usernameToSave string
	err := json.NewDecoder(r.Body).Decode(&usernameToSave)
	if err != nil {
		logrus.Errorf("cannot decode user name to save: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)

	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)

	client := client.NewClient(config.GimletD.URL, auth)
	createdUser, err := client.UserPost(&model.User{Login: usernameToSave})
	if err != nil {
		logrus.Errorf("cannot save GimletD user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	createdUserString, err := json.Marshal(createdUser)
	if err != nil {
		logrus.Errorf("cannot serialize user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(createdUserString)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)

	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)

	client := client.NewClient(config.GimletD.URL, auth)
	gimletdUsers, err := client.UsersGet()
	if err != nil {
		logrus.Errorf("cannot get GimletD users: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	usersString, err := json.Marshal(gimletdUsers)
	if err != nil {
		logrus.Errorf("cannot serialize users: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(usersString)
}

type App struct {
	Name     string        `json:"name"`
	Releases []*dx.Release `json:"releases"`
}

type Env struct {
	Name string `json:"name"`
	Apps []*App `json:"apps"`
}

func rolloutHistoryPerApp(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	env := chi.URLParam(r, "env")
	app := chi.URLParam(r, "app")
	repoName := fmt.Sprintf("%s/%s", owner, name)
	const perAppLimit = 10

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)

	// If GimletD is not set up, throw 404
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}

	releases, err := getAppReleasesFromGimletD(
		config.GimletD.URL,
		config.GimletD.TOKEN,
		config.ReleaseHistorySinceDays,
		perAppLimit,
		env,
		app,
		repoName,
	)
	if err != nil {
		logrus.Errorf("cannot get releases for git repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sort.Sort(ByCreated(releases))

	releasesString, err := json.Marshal(releases)
	if err != nil {
		logrus.Errorf("cannot serialize releases: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(releasesString)
}

func releaseStatuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	env := chi.URLParam(r, "env")
	perAppLimitString := r.URL.Query().Get("limit")

	perAppLimit, err := strconv.Atoi(perAppLimitString)
	if err != nil {
		logrus.Errorf("cannot convert limit: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// If GimletD is not set up, throw 404
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}

	releases, err := getAppReleasesFromGimletD(
		config.GimletD.URL,
		config.GimletD.TOKEN,
		config.ReleaseHistorySinceDays,
		perAppLimit,
		env,
		"",
		"",
	)
	if err != nil {
		logrus.Errorf("cannot get releases for git repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sort.Sort(ByCreated(releases))

	releasesString, err := json.Marshal(releases)
	if err != nil {
		logrus.Errorf("cannot serialize releases: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(releasesString)
}

func insertIntoRolloutHistory(rolloutHistory []*Env, release *dx.Release, perAppLimit int) []*Env {
	for _, env := range rolloutHistory {
		if env.Name == release.Env {
			for _, app := range env.Apps {
				if app.Name == release.App {
					if len(app.Releases) < perAppLimit {
						app.Releases = append(app.Releases, release)
					}
					return rolloutHistory
				}
			}

			env.Apps = append(env.Apps, &App{
				Name:     release.App,
				Releases: []*dx.Release{release},
			})
			return rolloutHistory
		}
	}

	rolloutHistory = append(rolloutHistory, &Env{
		Name: release.Env,
		Apps: []*App{
			{
				Name:     release.App,
				Releases: []*dx.Release{release},
			},
		},
	})

	return rolloutHistory
}

type ByCreated []*dx.Release

func (a ByCreated) Len() int           { return len(a) }
func (a ByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreated) Less(i, j int) bool { return a[i].Created < a[j].Created }

func orderRolloutHistoryFromAscending(rolloutHistory []*Env) []*Env {
	orderedRolloutHistory := []*Env{}

	for _, env := range rolloutHistory {
		orderedApps := []*App{}

		for _, app := range env.Apps {
			sort.Sort(ByCreated(app.Releases))
			orderedApps = append(orderedApps, app)
		}

		env.Apps = orderedApps
		orderedRolloutHistory = append(orderedRolloutHistory, env)
	}

	return orderedRolloutHistory
}

func gatherEnvsFromAgents(agentHub *streaming.AgentHub) []*api.ConnectedAgent {
	envs := []*api.ConnectedAgent{}
	for _, a := range agentHub.Agents {
		for _, stack := range a.Stacks {
			stack.Env = a.Name
		}
		envs = append(envs, &api.ConnectedAgent{
			Name:   a.Name,
			Stacks: a.Stacks,
		})
	}
	return envs
}

func getAppReleasesFromGimletD(
	gimletdURL string,
	gimletdToken string,
	releaseHistorySinceDays int,
	limit int,
	env string,
	app string,
	repoName string,
) ([]*dx.Release, error) {
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: gimletdToken,
		},
	)
	client := client.NewClient(gimletdURL, auth)

	// limiting query scope
	// without these, for apps released just once, the whole history would be traversed
	since := time.Now().Add(-1 * time.Hour * 24 * time.Duration(releaseHistorySinceDays))

	return client.ReleasesGet(
		app,
		env,
		limit,
		0,
		repoName,
		&since, nil,
	)
}

func deploy(w http.ResponseWriter, r *http.Request) {
	var releaseRequest dx.ReleaseRequest
	err := json.NewDecoder(r.Body).Decode(&releaseRequest)
	if err != nil {
		logrus.Errorf("cannot decode release request: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)
	adminClient := client.NewClient(config.GimletD.URL, auth)

	user := ctx.Value("user").(*model.User)
	gimletdUser, err := adminClient.UserGet(user.Login, true)
	if err != nil {
		logrus.Errorf("cannot find gimletd user: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	oauth2Config = new(oauth2.Config)
	auth = oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: gimletdUser.Token,
		},
	)
	impersonatedClient := client.NewClient(config.GimletD.URL, auth)

	trackingID, err := impersonatedClient.ReleasesPost(releaseRequest)
	if err != nil {
		logrus.Errorf("cannot post release: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	trackingString, err := json.Marshal(map[string]interface{}{
		"trackingId": trackingID,
	})
	if err != nil {
		logrus.Errorf("cannot serialize trackingId: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(trackingString)
}

func rollback(w http.ResponseWriter, r *http.Request) {
	var rollbackRequest dx.RollbackRequest
	err := json.NewDecoder(r.Body).Decode(&rollbackRequest)
	if err != nil {
		logrus.Errorf("cannot decode rollback request: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)
	adminClient := client.NewClient(config.GimletD.URL, auth)

	user := ctx.Value("user").(*model.User)
	gimletdUser, err := adminClient.UserGet(user.Login, true)
	if err != nil {
		logrus.Errorf("cannot find gimletd user: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	oauth2Config = new(oauth2.Config)
	auth = oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: gimletdUser.Token,
		},
	)
	impersonatedClient := client.NewClient(config.GimletD.URL, auth)

	trackingID, err := impersonatedClient.RollbackPost(
		rollbackRequest.Env,
		rollbackRequest.App,
		rollbackRequest.TargetSHA,
	)
	if err != nil {
		logrus.Errorf("cannot post rollback: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	trackingString, err := json.Marshal(map[string]interface{}{
		"trackingId": trackingID,
	})
	if err != nil {
		logrus.Errorf("cannot serialize trackingId: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(trackingString)
}

func deployStatus(w http.ResponseWriter, r *http.Request) {
	trackingId := r.URL.Query().Get("trackingId")
	if trackingId == "" {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), "trackingId parameter is mandatory"), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)
	client := client.NewClient(config.GimletD.URL, auth)

	releaseStatus, err := client.TrackRelease(trackingId)
	if err != nil {
		logrus.Errorf("cannot get deployStatus: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	releaseStatusString, err := json.Marshal(releaseStatus)
	if err != nil {
		logrus.Errorf("cannot serialize releaseStatus: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(releaseStatusString)
}

func getGitopsCommits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)
	client := client.NewClient(config.GimletD.URL, auth)

	gitopsCommits, err := client.GitopsCommitsGet(config.GimletD.TOKEN)
	if err != nil {
		logrus.Errorf("cannot get gitops commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	gitopsCommitsString, err := json.Marshal(gitopsCommits)
	if err != nil {
		logrus.Errorf("cannot serialize gitopsCommits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(gitopsCommitsString)
}

func decorateCommitsWithGimletArtifacts(commits []*Commit, config *config.Config) ([]*Commit, error) {
	if config.GimletD.URL == "" ||
		config.GimletD.TOKEN == "" {
		logrus.Warnf("couldn't connect to Gimletd for artifact data: gimletd access not configured")
		return commits, nil
	}
	oauth2Config := new(oauth2.Config)
	auth := oauth2Config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: config.GimletD.TOKEN,
		},
	)
	client := client.NewClient(config.GimletD.URL, auth)

	var hashes []string
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	artifacts, err := client.ArtifactsGet(
		"", "",
		nil,
		"",
		hashes,
		0, 0,
		nil, nil,
	)
	if err != nil {
		return commits, fmt.Errorf("cannot get artifacts: %s", err)
	}

	artifactsBySha := map[string]*dx.Artifact{}
	for _, a := range artifacts {
		artifactsBySha[a.Version.SHA] = a
	}

	var decoratedCommits []*Commit
	for _, c := range commits {
		if artifact, ok := artifactsBySha[c.SHA]; ok {
			for _, targetEnv := range artifact.Environments {
				targetEnv.ResolveVars(artifact.CollectVariables())
				if c.DeployTargets == nil {
					c.DeployTargets = []*DeployTarget{}
				}
				c.DeployTargets = append(c.DeployTargets, &DeployTarget{
					App:        targetEnv.App,
					Env:        targetEnv.Env,
					ArtifactId: artifact.ID,
				})
			}
		}
		decoratedCommits = append(decoratedCommits, c)
	}

	return decoratedCommits, nil
}
