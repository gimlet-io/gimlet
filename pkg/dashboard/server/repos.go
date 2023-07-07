package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
)

func gitRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)

	dao := ctx.Value("store").(*store.Store)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)

	userHasAccessToRepos, err := fetchReposWithAccess(dynamicConfig, tokenManager, dao, user)
	if err != nil {
		logrus.Errorf("cannot get repos from db: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
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

func refreshRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)

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
	updatedUserRepos := updateUserRepos(dynamicConfig, dao, user)
	updatedOrgRepos := updateOrgRepos(dynamicConfig, tokenManager, dao)
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

func updateOrgRepos(
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
	dao *store.Store,
) []string {
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
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

func updateUserRepos(dynamicConfig *dynamicconfig.DynamicConfig, dao *store.Store, user *model.User) []string {
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, func(token *scm.Token) {
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

func fetchReposWithAccess(
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
	dao *store.Store,
	user *model.User,
) ([]string, error) {
	userRepos := user.Repos

	orgRepos, err := getOrgRepos(dao)
	if err != nil {
		return nil, err
	}

	if len(userRepos) == 0 && len(orgRepos) == 0 {
		userRepos = updateUserRepos(dynamicConfig, dao, user)
		orgRepos = updateOrgRepos(dynamicConfig, tokenManager, dao)
	}

	return intersection(orgRepos, userRepos), nil
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
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)

	var provider string
	if dynamicConfig.IsGithub() {
		provider = "github"
	} else if dynamicConfig.IsGitlab() {
		provider = "gitlab"
	}

	settings := map[string]interface{}{
		"releaseHistorySinceDays": config.ReleaseHistorySinceDays,
		"scmUrl":                  dynamicConfig.ScmURL(),
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
