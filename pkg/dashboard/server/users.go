package server

import (
	"encoding/json"
	"net/http"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
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

	onlyAPIKeys := []*model.User{}
	for _, u := range users {
		if !u.Admin {
			onlyAPIKeys = append(onlyAPIKeys, u)
		}
	}

	usersString, err := json.Marshal(onlyAPIKeys)
	if err != nil {
		logrus.Errorf("cannot serialize users: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(usersString)
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
