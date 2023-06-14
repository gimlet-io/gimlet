package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func gitRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)

	dao := ctx.Value("store").(*store.Store)
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)

	go updateUserRepos(config, dao, user)

	timeout := time.After(45 * time.Second)
	user, err := func() (*model.User, error) {
		for {
			user, err := dao.User(user.Login)
			if err != nil {
				logrus.Errorf("cannot get user from db: %s", err)
				return nil, err
			}

			if len(user.Repos) > 0 {
				return user, nil
			}

			select {
			case <-timeout:
				return &model.User{}, nil
			default:
				time.Sleep(3 * time.Second)
			}
		}
	}()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userHasAccessToRepos := hasPrefix(user.Repos, config.Org())
	reposString, err := json.Marshal(userHasAccessToRepos)
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(reposString)
}

func refreshRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)

	user, err := dao.User(user.Login)
	if err != nil {
		logrus.Errorf("cannot get user from db: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userReposDb := hasPrefix(user.Repos, config.Org())
	userRepos := updateUserRepos(config, dao, user)
	userReposWithAccess := hasPrefix(userRepos, config.Org())
	added := difference(userReposWithAccess, userReposDb)
	deleted := difference(userReposDb, userReposWithAccess)

	repos := map[string]interface{}{
		"userRepos": userReposWithAccess,
		"added":     added,
		"deleted":   deleted,
	}

	reposString, err := json.Marshal(repos)
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(reposString)
}

func updateOrgRepos(ctx context.Context) {
	gitServiceImpl := *ctx.Value("gitService").(*customScm.CustomGitService)
	tokenManager := *ctx.Value("tokenManager").(*customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	orgRepos, err := gitServiceImpl.OrgRepos(token)
	if err != nil {
		logrus.Warnf("cannot get org repos: %s", err)
		return
	}

	orgReposString, err := json.Marshal(orgRepos)
	if err != nil {
		logrus.Warnf("cannot serialize org repos: %s", err)
		return
	}

	dao := ctx.Value("store").(*store.Store)
	err = dao.SaveKeyValue(&model.KeyValue{
		Key:   model.OrgRepos,
		Value: string(orgReposString),
	})
	if err != nil {
		logrus.Warnf("cannot store org repos: %s", err)
		return
	}
}

func updateUserRepos(config *config.PersistentConfig, dao *store.Store, user *model.User) []string {
	goScmHelper := genericScm.NewGoScmHelper(config, func(token *scm.Token) {
		user.AccessToken = token.Token
		user.RefreshToken = token.Refresh
		user.Expires = token.Expires.Unix()
		err := dao.UpdateUser(user)
		if err != nil {
			logrus.Errorf("could not refresh user's oauth access_token")
		}
	})
	userRepos, err := goScmHelper.UserRepos(user.AccessToken, user.RefreshToken, time.Unix(user.Expires, 0))
	if err != nil {
		logrus.Warnf("cannot get user repos: %s", err)
		return nil
	}

	user.Repos = userRepos
	err = dao.UpdateUser(user)
	if err != nil {
		logrus.Warnf("cannot get user repos: %s", err)
		return nil
	}
	return userRepos
}

func hasPrefix(repos []string, prefix string) []string {
	filtered := []string{}
	for _, r := range repos {
		if strings.HasPrefix(r, prefix) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

type favRepos struct {
	FavoriteRepos []string `json:"favoriteRepos"`
}

func saveFavoriteRepos(w http.ResponseWriter, r *http.Request) {
	var reposPayload favRepos
	err := json.NewDecoder(r.Body).Decode(&reposPayload)
	if err != nil {
		logrus.Errorf("cannot decode repos payload: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)

	user.FavoriteRepos = reposPayload.FavoriteRepos
	err = dao.UpdateUser(user)
	if err != nil {
		logrus.Errorf("cannot save favorite repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("{}"))
}

func saveFavoriteServices(w http.ResponseWriter, r *http.Request) {
	var servicesPayload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&servicesPayload)
	if err != nil {
		logrus.Errorf("cannot decode services payload: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	var services []string
	if s, ok := servicesPayload["favoriteServices"]; ok {
		services = s.([]string)
	} else {
		logrus.Errorf("cannot get favorite services: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)

	user.FavoriteServices = services
	err = dao.UpdateUser(user)
	if err != nil {
		logrus.Errorf("cannot save favorite services: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("{}"))
}

func settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("persistentConfig").(*config.PersistentConfig)

	var provider string
	if config.IsGithub() {
		provider = "github"
	} else if config.IsGitlab() {
		provider = "gitlab"
	}

	settings := map[string]interface{}{
		"releaseHistorySinceDays": config.Get(store.ReleaseHistorySinceDays),
		"scmUrl":                  config.ScmURL(),
		"userflowToken":           config.Get(store.UserflowToken),
		"host":                    config.Get(store.Host),
		"provider":                provider,
	}

	settingsString, err := json.Marshal(settings)
	if err != nil {
		logrus.Errorf("cannot serialize settings: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(settingsString))
}

// returns the elements in `slice1` that aren't in `slice2`
func difference(slice1 []string, slice2 []string) []string {
	mb := make(map[string]struct{}, len(slice2))
	for _, x := range slice2 {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range slice1 {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
