package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
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

	var orgRepos []string
	dao := ctx.Value("store").(*store.Store)
	config := ctx.Value("config").(*config.Config)

	go updateOrgRepos(ctx)
	go updateUserRepos(config, dao, user)

	timeout := time.After(45 * time.Second)
	orgReposJson, user, err := func() (*model.KeyValue, *model.User, error) {
		for {
			orgReposJson, err := dao.KeyValue(model.OrgRepos)
			if err != nil && err != sql.ErrNoRows {
				logrus.Errorf("cannot load org repos: %s", err)
				return nil, nil, err
			}

			user, err = dao.User(user.Login)
			if err != nil {
				logrus.Errorf("cannot get user from db: %s", err)
				return nil, nil, err
			}

			if orgReposJson.Value != "" && len(user.Repos) > 0 {
				return orgReposJson, user, nil
			}

			select {
			case <-timeout:
				return &model.KeyValue{}, &model.User{}, nil
			default:
				time.Sleep(3 * time.Second)
			}
		}
	}()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	if orgReposJson.Value == "" {
		orgReposJson.Value = "[]"
	}

	err = json.Unmarshal([]byte(orgReposJson.Value), &orgRepos)
	if err != nil {
		logrus.Errorf("cannot unmarshal org repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userHasAccessToRepos := intersection(orgRepos, user.Repos)
	if userHasAccessToRepos == nil {
		userHasAccessToRepos = []string{}
	}
	reposString, err := json.Marshal(userHasAccessToRepos)
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(reposString)
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

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}

	return
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
