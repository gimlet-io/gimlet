package server

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/httputil"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet/pkg/server/token"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/gorilla/securecookie"
	"github.com/laszlocph/go-login/login"
	log "github.com/sirupsen/logrus"
)

func auth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := login.ErrorFrom(ctx)
	if err != nil {
		log.Errorf("cannot get access token: %s", err)
		http.Error(w, "Cannot decode token", 400)
		return
	}
	token := login.TokenFrom(ctx)

	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, nil)
	scmUser, err := goScmHelper.User(token.Access, token.Refresh)
	if err != nil {
		log.Errorf("cannot find git user: %s", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	orgList, err := goScmHelper.Organizations(token.Access, token.Refresh)
	if err != nil {
		log.Errorf("cannot get user organizations: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	member := validateOrganizationMembership(orgList, dynamicConfig.Org(), scmUser.Login)
	if !member {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	store := ctx.Value("store").(*store.Store)
	user, err := getOrCreateUser(store, scmUser, token)
	if err != nil {
		log.Errorf("cannot get or store user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	_, err = updateUserRepos(dynamicConfig, store, user)
	if err != nil {
		log.Errorf("cannot update user repos: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = setSessionCookie(w, r, user)
	if err != nil {
		log.Errorf("cannot set session cookie: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.RedirectHandler(redirectPath(r.FormValue("state")), http.StatusSeeOther).ServeHTTP(w, r)
}

func redirectPath(value string) string {
	redirect := "/"
	parts := strings.Split(value, "&")
	if len(parts) != 3 {
		return redirect
	}

	params, _ := url.ParseQuery(parts[2])
	if v, found := params["redirect"]; found {
		redirect = v[0]
	}
	return redirect
}

func adminKeyAuth(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.Form
	adminKey := formValues.Get("token")
	ctx := r.Context()
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)

	if adminKey != dynamicConfig.AdminKey {
		log.Errorf("token is not valid")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	store := ctx.Value("store").(*store.Store)
	admin, err := store.User("admin")
	if err != nil {
		log.Errorf("cannot get user from db: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = setSessionCookie(w, r, admin)
	if err != nil {
		log.Errorf("cannot set session cookie: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func validateOrganizationMembership(orgList []*scm.Organization, org string, userName string) bool {
	if org == userName { // allowing single user installations
		return true
	}

	for _, organization := range orgList {
		if org == organization.Name {
			return true
		}
	}
	return false
}

func logout(w http.ResponseWriter, r *http.Request) {
	httputil.DelCookie(w, r, "user_sess")
	http.RedirectHandler("/login", http.StatusSeeOther).ServeHTTP(w, r)
}

func getOrCreateUser(store *store.Store, scmUser *scm.User, token *login.Token) (*model.User, error) {
	user, err := store.User(scmUser.Login)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			user = &model.User{
				Login:        scmUser.Login,
				Name:         scmUser.Name,
				Email:        scmUser.Email,
				Admin:        true,
				AccessToken:  token.Access,
				RefreshToken: token.Refresh,
				Expires:      token.Expires.Unix(),
				Secret: base32.StdEncoding.EncodeToString(
					securecookie.GenerateRandomKey(32),
				),
				FavoriteRepos:    []string{},
				FavoriteServices: []string{},
			}
			err = store.CreateUser(user)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	} else {
		user.Name = scmUser.Name // Remove this 2 releases from now
		user.AccessToken = token.Access
		user.RefreshToken = token.Refresh
		user.Expires = token.Expires.Unix()
		err = store.UpdateUser(user)
		if err != nil {
			return nil, err
		}
	}

	return user, err
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, user *model.User) error {
	fortyEightHours, _ := time.ParseDuration("48h")
	exp := time.Now().Add(fortyEightHours).Unix()
	t := token.New(token.SessToken, user.Login, "")
	tokenStr, err := t.SignExpires(user.Secret, exp)
	if err != nil {
		return err
	}

	httputil.SetCookie(w, r, "user_sess", tokenStr)
	return nil
}
