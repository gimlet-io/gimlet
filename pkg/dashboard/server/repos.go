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
	config := ctx.Value("config").(*config.Config)

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
	config := ctx.Value("config").(*config.Config)

	user, err := dao.User(user.Login)
	if err != nil {
		logrus.Errorf("cannot get user from db: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	go updateUserRepos(config, dao, user)

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
		return
	}

	userHasAccessToReposDb := hasPrefix(user.Repos, config.Org())
	userHasAccessToReposGit := hasPrefix(userRepos, config.Org())
	repoDiffs := diff(userHasAccessToReposDb, userHasAccessToReposGit)
	added, deleted := repoDifferences(userHasAccessToReposDb, userHasAccessToReposGit, repoDiffs)

	refresh := map[string]interface{}{
		"repos":   userHasAccessToReposGit,
		"added":   added,
		"deleted": deleted,
	}

	refreshString, err := json.Marshal(refresh)
	if err != nil {
		logrus.Errorf("cannot serialize refresh: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(refreshString)
}

func updateOrgRepos(ctx context.Context) {
	gitServiceImpl := ctx.Value("gitService").(customScm.CustomGitService)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
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

func updateUserRepos(config *config.Config, dao *store.Store, user *model.User) {
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
		return
	}

	user.Repos = userRepos
	err = dao.UpdateUser(user)
	if err != nil {
		logrus.Warnf("cannot get user repos: %s", err)
		return
	}
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
	config := ctx.Value("config").(*config.Config)
	settings := map[string]interface{}{
		"releaseHistorySinceDays": config.ReleaseHistorySinceDays,
		"scmUrl":                  config.ScmURL(),
		"userflowToken":           config.UserflowToken,
		"host":                    config.Host,
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

func repoDifferences(dbRepos, gitRepos, repoDiffs []string) (added, deleted []string) {
	added = make([]string, 0, len(repoDiffs))
	deleted = make([]string, 0, len(repoDiffs))
	for _, repo := range repoDiffs {
		if contains(dbRepos, repo) {
			deleted = append(deleted, repo)
		}

		if contains(gitRepos, repo) {
			added = append(added, repo)
		}
	}

	return added, deleted
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func diff(slice1 []string, slice2 []string) []string {
	var diff []string

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
