package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/genericScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-chi/chi/v5"
	"github.com/hyperboloide/lk"
	"github.com/sirupsen/logrus"
)

func gitRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)

	importedRepos, err := getImportedRepos(dao)
	if err != nil {
		logrus.Errorf("cannot get user repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	userHasAccessToRepos := intersection(importedRepos, user.Repos)

	reposString, err := json.Marshal(userHasAccessToRepos)
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(reposString)
}

func searchRepo(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)

	var userRepos []string
	var err error
	if query == "" {
		userRepos = user.Repos
	} else {
		userRepos, err = updateUserRepos(dynamicConfig, dao, user)
		if err != nil {
			logrus.Errorf("cannot get user repos: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
	}

	gitServiceImpl := customScm.NewGitService(dynamicConfig)
	token, _, _ := tokenManager.Token()
	installationRepos, err := gitServiceImpl.InstallationRepos(token)
	if err != nil {
		logrus.Errorf("cannot get installation repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	filteredRepos := []string{}
	for _, repo := range intersection(installationRepos, userRepos) {
		if strings.Contains(strings.ToLower(repo), strings.ToLower(query)) {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	filteredReposString, err := json.Marshal(filteredRepos)
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(filteredReposString)
}

func importRepo(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("name")
	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	dao := ctx.Value("store").(*store.Store)

	importedRepos, err := updateImportedReposWithRepo(dao, repoName)
	if err != nil {
		logrus.Errorf("cannot update imported repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	reposString, err := json.Marshal(intersection(importedRepos, user.Repos))
	if err != nil {
		logrus.Errorf("cannot serialize repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(reposString))
}

func updateImportedReposWithRepo(dao *store.Store, repoName string) ([]string, error) {
	importedRepos, err := getImportedRepos(dao)
	if err != nil {
		return nil, err
	}

	importedRepos = appendIfMissing(importedRepos, repoName)
	importedReposString, err := json.Marshal(importedRepos)
	if err != nil {
		return nil, err
	}

	err = dao.SaveKeyValue(&model.KeyValue{
		Key:   model.ImportedRepos,
		Value: string(importedReposString),
	})
	if err != nil {
		return nil, err
	}
	return importedRepos, nil
}

func appendIfMissing(slice []string, i string) []string {
	for _, e := range slice {
		if e == i {
			return slice
		}
	}
	return append(slice, i)
}

func updateUserRepos(dynamicConfig *dynamicconfig.DynamicConfig, dao *store.Store, user *model.User) ([]string, error) {
	userRepos, err := userReposFromGithub(dynamicConfig, dao, user)
	if err != nil {
		return nil, err
	}

	user.Repos = userRepos
	err = dao.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	return userRepos, nil
}

func userReposFromGithub(dynamicConfig *dynamicconfig.DynamicConfig, dao *store.Store, user *model.User) ([]string, error) {
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, func(token *scm.Token) {
		user.AccessToken = token.Token
		user.RefreshToken = token.Refresh
		user.Expires = token.Expires.Unix()
		err := dao.UpdateUser(user)
		if err != nil {
			logrus.Errorf("could not refresh user's oauth access_token")
		}
	})

	return goScmHelper.UserRepos(user.AccessToken, user.RefreshToken, time.Unix(user.Expires, 0))
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

func repoPullRequestPolicy(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "name")

	ctx := r.Context()
	dao := ctx.Value("store").(*store.Store)
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	pullRequestPolicy, err := dao.RepoHasPullRequestPolicy(repoName)
	if err != nil {
		logrus.Errorf("cannot get pull request policy for %s: %s", repoName, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	metas := gitRepoMetas{
		PullRequestPolicy: pullRequestPolicy,
	}

	metasString, err := json.Marshal(metas)
	if err != nil {
		logrus.Errorf("cannot serialize repo meta: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(metasString)
}

type pullRequestPolicy struct {
	PullRequestPolicy bool `json:"pullRequestPolicy"`
}

func saveRepoPullRequestPolicy(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := chi.URLParam(r, "name")
	repoPath := fmt.Sprintf("%s/%s", owner, repoName)

	var payload pullRequestPolicy
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		logrus.Errorf("cannot decode repos payload: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	dao := ctx.Value("store").(*store.Store)
	err = dao.SaveRepoPullRequestPolicy(repoPath, payload.PullRequestPolicy)
	if err != nil {
		logrus.Errorf("could not save repo pull request policy: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config := ctx.Value("config").(*config.Config)
	dao := ctx.Value("store").(*store.Store)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)

	var provider string
	if dynamicConfig.IsGithub() {
		provider = "github"
	} else if dynamicConfig.IsGitlab() {
		provider = "gitlab"
	}

	activatedTrial := false
	_, err := dao.KeyValue(model.ActivatedTrial)
	if err == nil { // trial is activated
		activatedTrial = true
	}

	settings := map[string]interface{}{
		"releaseHistorySinceDays": config.ReleaseHistorySinceDays,
		"posthogFeatureFlag":      config.PosthogFeatureFlag(),
		"posthogIdentifyUser":     config.PosthogIdentifyUser,
		"posthogApiKey":           config.PosthogApiKey,
		"scmUrl":                  dynamicConfig.ScmURL(),
		"host":                    config.Host,
		"provider":                provider,
		"trial":                   config.Instance != "" && !activatedTrial,
		"instance":                config.Instance,
	}

	licensed, err := validateLicense(
		config.License,
		"ARO2ZN3OVUQT6QWSUATB2JGQYZFXZ2JVGKHU5SZGRIDF5EJ5IL4UJORP6UM23QBMMQPLRB6KE467QIDIIGAMNWP5KGMOMRWDLCGOEKK2BRILYFURXFL37NH5AJZ4QSXU7VL5SIYVHXBZ24YHZ3EUE7GXVNXA====",
	)
	if err == nil {
		settings["licensed"] = licensed
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

type ValidLicense struct {
	Name string `json:"name"`
	// Email string    `json:"email"`
	End time.Time `json:"until"`
}

func validateLicense(licenseB32, publicKeyBase32 string) (*ValidLicense, error) {
	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB32String(publicKeyBase32)
	if err != nil {
		return nil, err
	}

	// Unmarshal the customer license.
	license, err := lk.LicenseFromB32String(licenseB32)
	if err != nil {
		return nil, err
	}

	// validate the license signature.
	if ok, err := license.Verify(publicKey); err != nil {
		return nil, err
	} else if !ok {
		return nil, err
	}

	// unmarshal the document.
	var licensed ValidLicense
	if err := json.Unmarshal(license.Data, &licensed); err != nil {
		return nil, err
	}

	// Now you just have to check that the end date is after time.Now() then you can continue!
	if licensed.End.Before(time.Now()) {
		return nil, fmt.Errorf("license expired on: %s", licensed.End.Format("2006-01-02"))
	}

	return &licensed, nil
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
