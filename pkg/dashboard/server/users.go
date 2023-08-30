package server

import (
	"bytes"
	"encoding/base32"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/go-chi/chi"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	users, err := store.Users()
	if err != nil {
		logrus.Errorf("cannot get users: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	usersString, err := json.Marshal(users)
	if err != nil {
		logrus.Errorf("cannot serialize users: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(usersString)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	login := chi.URLParam(r, "login")
	user, err := store.User(login)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		logrus.Errorf("cannot get user %s: %s", login, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	withToken := r.URL.Query().Get("withToken")
	if withToken == "true" {
		token := token.New(token.UserToken, user.Login)
		tokenStr, err := token.Sign(user.Secret)
		if err != nil {
			logrus.Errorf("couldn't generate JWT token %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		// token is not saved as it is JWT
		user.Token = tokenStr
	}

	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user %s: %s", login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userString)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var usernameToDelete string
	err := json.NewDecoder(r.Body).Decode(&usernameToDelete)
	if err != nil {
		logrus.Errorf("cannot decode user name to delete: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	user := ctx.Value("user").(*model.User)
	if usernameToDelete == user.Login {
		logrus.Errorf("self-deletion is not allowed")
		http.Error(w, http.StatusText(400), 400)
		return
	}

	store := ctx.Value("store").(*store.Store)

	err = store.DeleteUser(usernameToDelete)
	if err != nil {
		logrus.Errorf("cannot delete user %s: %s", usernameToDelete, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func saveUserGimletD(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var user model.User
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&user)
	if err != nil {
		logrus.Errorf("cannot decode user: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	user.Secret = base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	err = store.CreateUser(&user)
	if err != nil {
		logrus.Errorf("cannot creat user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	token := token.New(token.UserToken, user.Login)
	tokenStr, err := token.Sign(user.Secret)
	if err != nil {
		logrus.Errorf("couldn't create user token %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// token is not saved as it is JWT
	user.Token = tokenStr

	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(userString)
}
