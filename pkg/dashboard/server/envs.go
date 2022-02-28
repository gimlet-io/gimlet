package server

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func saveInfrastructureComponents(w http.ResponseWriter, r *http.Request) {
	var infrastructureComponents map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&infrastructureComponents)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	infrastructureComponentsString, err := json.Marshal(infrastructureComponents)
	if err != nil {
		logrus.Errorf("cannot serialize infrastructure components: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(infrastructureComponentsString)
}
