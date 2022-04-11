package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/session"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
)

func SetupRouter(
	config *config.Config,
	store *store.Store,
	notificationsManager notifications.Manager,
	repoCache *nativeGit.GitopsRepoCache,
	perf *prometheus.HistogramVec,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.WithValue("notificationsManager", notificationsManager))
	r.Use(middleware.WithValue("gitopsRepo", config.GitopsRepo))
	r.Use(middleware.WithValue("gitopsRepoDeployKeyPath", config.GitopsRepoDeployKeyPath))
	r.Use(middleware.WithValue("gitopsRepoCache", repoCache))
	r.Use(middleware.WithValue("perf", perf))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8888", config.Host},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.MustUser())
		r.Post("/api/artifact", saveArtifact)
		r.Get("/api/artifacts", getArtifacts)
		r.Get("/api/releases", getReleases)
		r.Get("/api/status", getStatus)
		r.Post("/api/releases", release)
		r.Post("/api/rollback", rollback)
		r.Post("/api/delete", delete)
		r.Get("/api/event", getEvent)
		r.Post("/api/flux-events", fluxEvent)
		r.Get("/api/gitopsCommits", getGitopsCommits)

		r.Get("/api/gitopsRepo", func(w http.ResponseWriter, r *http.Request) {
			gitopsRepo := r.Context().Value("gitopsRepo").(string)
			gitopsRepoJson, _ := json.Marshal(GitopsRepoResult{GitopsRepo: gitopsRepo})
			w.WriteHeader(http.StatusOK)
			w.Write(gitopsRepoJson)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.MustAdmin())
		r.Get("/api/user/{login}", getUser)
		r.Post("/api/user", saveUser)
		r.Delete("/api/user/{login}", deleteUser)
		r.Get("/api/users", getUsers)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

type GitopsRepoResult struct {
	GitopsRepo string `json:"gitopsRepo"`
}
