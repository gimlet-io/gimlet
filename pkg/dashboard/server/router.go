package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/alert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/session"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/laszlocph/go-login/login/logger"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/laszlocph/go-login/login/github"
	"github.com/laszlocph/go-login/login/gitlab"
	log "github.com/sirupsen/logrus"
)

var agentAuth *jwtauth.JWTAuth

func SetupRouter(
	config *config.Config,
	agentHub *streaming.AgentHub,
	clientHub *streaming.ClientHub,
	agentWSHub *streaming.AgentWSHub,
	store *store.Store,
	gitService customScm.CustomGitService,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *nativeGit.RepoCache,
	chartUpdatePullRequests *map[string]interface{},
	alertStateManager *alert.AlertStateManager,
	notificationsManager notifications.Manager,
	perf *prometheus.HistogramVec,
	logger *log.Logger,
) *chi.Mux {
	agentAuth = jwtauth.New("HS256", []byte(config.JWTSecret), nil)
	_, tokenString, _ := agentAuth.Encode(map[string]interface{}{"user_id": "gimlet-agent"})
	log.Infof("Agent JWT is %s\n", tokenString)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(middleware.WithValue("agentHub", agentHub))
	r.Use(middleware.WithValue("clientHub", clientHub))
	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.WithValue("config", config))
	r.Use(middleware.WithValue("gitService", gitService))
	r.Use(middleware.WithValue("tokenManager", tokenManager))
	r.Use(middleware.WithValue("gitRepoCache", repoCache))
	r.Use(middleware.WithValue("agentJWT", tokenString))
	r.Use(middleware.WithValue("alertStateManager", alertStateManager))
	r.Use(middleware.WithValue("chartUpdatePullRequests", chartUpdatePullRequests))

	r.Use(middleware.WithValue("notificationsManager", notificationsManager))
	r.Use(middleware.WithValue("perf", perf))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:9000", "http://127.0.0.1:9000", config.Host},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	agentRoutes(r, agentWSHub)
	userRoutes(r)
	githubOAuthRoutes(config, r)
	gimletdRoutes(r)

	r.Get("/logout", logout)

	r.Get("/ws/", func(w http.ResponseWriter, r *http.Request) {
		streaming.ServeWs(clientHub, w, r)
	})

	r.Post("/hook", hook)

	r.Get("/created", created)
	r.Get("/installed", installed)

	r.Get("/flags", getFlags)

	filesDir := http.Dir("./web/build")
	fileServer(r, "/", filesDir)
	fileServer(r, "/login", filesDir)
	fileServer(r, "/repositories", filesDir)
	fileServer(r, "/pulse", filesDir)
	fileServer(r, "/profile", filesDir)
	fileServer(r, "/settings", filesDir)
	fileServer(r, "/repo", filesDir)
	fileServer(r, "/environments", filesDir)

	return r
}

func gimletdRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.MustUser())
		r.Post("/api/artifact", saveArtifact)
		r.Get("/api/artifacts", getArtifacts)
		r.Get("/api/releases", getReleases)
		r.Get("/api/status", getStatus)
		r.Post("/api/releases", release)
		r.Post("/api/rollback", performRollback)
		r.Post("/api/delete", delete)
		r.Get("/api/eventReleaseTrack", getEventReleaseTrack)
		r.Get("/api/eventArtifactTrack", getEventArtifactTrack)
		r.Post("/api/flux-events", fluxEvent)
		r.Get("/api/gitopsCommits", getGitopsCommits)
	})

	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.MustAdmin())
		r.Get("/api/user/{login}", getUser)
		r.Post("/api/user", saveUserGimletD)
		r.Delete("/api/user/{login}", deleteUser)
		r.Get("/api/users", getUsers)
	})
}

func userRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.SetCSRF())
		r.Use(session.MustUser())

		r.Get("/api/agents", agents)
		r.Get("/api/user", user)
		r.Post("/api/saveUser", saveUser)
		r.Get("/api/envs", envs)
		r.Get("/api/podLogs", getPodLogs)
		r.Get("/api/stopPodLogs", stopPodLogs)
		r.Get("/api/alerts", getAlerts)
		r.Get("/api/gitRepos", gitRepos)
		r.Get("/api/refreshRepos", refreshRepos)
		r.Get("/api/settings", settings)
		r.Get("/api/repo/{owner}/{name}/commits", commits)
		r.Get("/api/gitopsCommits", getGitopsCommits)
		r.Get("/api/repo/{owner}/{name}/branches", branches)
		r.Get("/api/repo/{owner}/{name}/metas", getMetas)
		r.Get("/api/repo/{owner}/{name}/pullRequests", getPullRequests)
		r.Get("/api/chartUpdatePullRequests", getChartUpdatePullRequests)
		r.Get("/api/infraRepoPullRequests", getPullRequestsFromInfraRepos)
		r.Get("/api/repo/{owner}/{name}/envConfigs", envConfigs)
		r.Post("/api/repo/{owner}/{name}/env/{env}/config/{config}", saveEnvConfig)
		r.Post("/api/saveFavoriteRepos", saveFavoriteRepos)
		r.Post("/api/saveFavoriteServices", saveFavoriteServices)
		r.Get("/api/repo/{owner}/{name}/env/{env}/chartSchema", chartSchema)
		r.Get(("/api/app"), application)
		r.Post(("/api/saveEnvToDB"), saveEnvToDB)
		r.Post(("/api/deleteEnvFromDB"), deleteEnvFromDB)
		r.Post(("/api/environments"), saveInfrastructureComponents)
		r.Post(("/api/bootstrapGitops"), bootstrapGitops)
		r.Post(("/api/envs/{env}/installAgent"), installAgent)
	})
}

func agentRoutes(r *chi.Mux, agentWSHub *streaming.AgentWSHub) {
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(agentAuth))
		r.Use(jwtauth.Authenticator)
		r.Use(mustAgent)

		r.Get("/agent/register", register)
		r.Post("/agent/state", state)
		r.Post("/agent/state/{name}/update", update)
		r.Post("/agent/events", events)

		r.Get("/agent/ws/", func(w http.ResponseWriter, r *http.Request) {
			streaming.ServeAgentWs(agentWSHub, w, r)
		})
	})
}

func githubOAuthRoutes(config *config.Config, r *chi.Mux) {
	if config.IsGithub() {
		dumper := logger.DiscardDumper()
		if config.Github.Debug {
			dumper = logger.StandardDumper()
		}
		loginMiddleware := &github.Config{
			ClientID:     config.Github.ClientID,
			ClientSecret: config.Github.ClientSecret,
			// you don't need to provide scopes in your authorization request.
			// Unlike traditional OAuth, the authorization token is limited to the permissions associated
			// with your GitHub App and those of the user.
			// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps#identifying-and-authorizing-users-for-github-apps
			Scope:  []string{""},
			Dumper: dumper,
		}

		r.Handle("/auth", loginMiddleware.Handler(
			http.HandlerFunc(auth),
		))
		r.Handle("/auth/*", loginMiddleware.Handler(
			http.HandlerFunc(auth),
		))
	} else if config.IsGitlab() {
		loginMiddleware := &gitlab.Config{
			Server:       config.ScmURL(),
			ClientID:     config.Gitlab.ClientID,
			ClientSecret: config.Gitlab.ClientSecret,
			RedirectURL:  config.Host + "/auth",
			Scope:        []string{"api"},
		}

		r.Handle("/auth", loginMiddleware.Handler(
			http.HandlerFunc(auth),
		))
		r.Handle("/auth/*", loginMiddleware.Handler(
			http.HandlerFunc(auth),
		))
	}
}

// static files from a http.FileSystem
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		//TODO: serve all React routes https://github.com/go-chi/chi/issues/403
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
		if pathPrefix == "/repo" ||
			pathPrefix == "/environments" {
			pathPrefix = r.URL.Path
		}
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func mustAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		userId := claims["user_id"]
		if userId != "gimlet-agent" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
