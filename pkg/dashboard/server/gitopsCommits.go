package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

func getGitopsCommits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	commits, err := store.GitopsCommits()
	if err == sql.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		logrus.Errorf("cannot get gitops commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	commitsString, err := json.Marshal(commits)
	if err != nil {
		logrus.Errorf("cannot serialize commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(commitsString)
}
