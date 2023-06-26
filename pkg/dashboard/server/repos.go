package server

import (
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

	dao := ctx.Value("store").(*store.Store)
	config := ctx.Value("config").(*config.Config)

	go updateUserRepos(config, dao, user)
	go updateOrgRepos(ctx)

	orgRepos, userRepos, err := fetchReposFromDb(dao, user.Login)
	if err != nil {
		logrus.Errorf("cannot get repos from db: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userHasAccessToRepos := intersection(orgRepos, userRepos)
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

	orgRepos, err := getOrgRepos(dao)
	if err != nil {
		logrus.Errorf("cannot get org repos from db: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userAccessRepos := intersection(orgRepos, user.Repos)
	updatedUserRepos := updateUserRepos(config, dao, user)
	updatedOrgRepos := updateOrgRepos(ctx)
	updatedUserAccessRepos := intersection(updatedOrgRepos, updatedUserRepos)
	added := difference(updatedUserAccessRepos, userAccessRepos)
	deleted := difference(userAccessRepos, updatedUserAccessRepos)

	repos := map[string]interface{}{
		"userRepos": updatedUserAccessRepos,
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

func updateOrgRepos(ctx context.Context) []string {
	config := ctx.Value("config").(*config.Config)
	gitServiceImpl := customScm.NewGitService(config)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()

	orgRepos, err := gitServiceImpl.OrgRepos(token)
	if err != nil {
		logrus.Warnf("cannot get org repos: %s", err)
		return nil
	}

	orgReposString, err := json.Marshal(orgRepos)
	if err != nil {
		logrus.Warnf("cannot serialize org repos: %s", err)
		return nil
	}

	dao := ctx.Value("store").(*store.Store)
	err = dao.SaveKeyValue(&model.KeyValue{
		Key:   model.OrgRepos,
		Value: string(orgReposString),
	})
	if err != nil {
		logrus.Warnf("cannot store org repos: %s", err)
		return nil
	}
	return orgRepos
}

func updateUserRepos(config *config.Config, dao *store.Store, user *model.User) []string {
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

func fetchReposFromDb(dao *store.Store, login string) ([]string, []string, error) {
	timeout := time.After(45 * time.Second)

	for {
		orgRepos, err := getOrgRepos(dao)
		if err != nil {
			return nil, nil, err
		}

		user, err := dao.User(login)
		if err != nil {
			return nil, nil, err
		}

		if len(orgRepos) != 0 && len(user.Repos) != 0 {
			return orgRepos, user.Repos, nil
		}

		select {
		case <-timeout:
			return orgRepos, user.Repos, nil
		default:
			time.Sleep(3 * time.Second)
		}
	}
}

func intersection(s1, s2 []string) []string {
	inter := []string{}
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

	return inter
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

	provider := "github"
	if config.IsGitlab() {
		provider = "gitlab"
	}

	settings := map[string]interface{}{
		"releaseHistorySinceDays": config.ReleaseHistorySinceDays,
		"scmUrl":                  config.ScmURL(),
		"userflowToken":           config.UserflowToken,
		"host":                    config.Host,
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
