package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/alert"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/session"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/go-chi/chi/v5"
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
	dynamicConfig *dynamicconfig.DynamicConfig,
	agentHub *streaming.AgentHub,
	clientHub *streaming.ClientHub,
	agentWSHub *streaming.AgentWSHub,
	store *store.Store,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *nativeGit.RepoCache,
	chartUpdatePullRequests *map[string]interface{},
	gitopsUpdatePullRequests *map[string]interface{},
	alertStateManager *alert.AlertStateManager,
	notificationsManager notifications.Manager,
	perf *prometheus.HistogramVec,
	logger *log.Logger,
	gitServer http.Handler,
	gitUser *model.User,
	gitopsQueue chan int,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)

	r.Use(middleware.WithValue("agentHub", agentHub))
	r.Use(middleware.WithValue("clientHub", clientHub))
	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.WithValue("config", config))
	r.Use(middleware.WithValue("dynamicConfig", dynamicConfig))
	r.Use(middleware.WithValue("tokenManager", tokenManager))
	r.Use(middleware.WithValue("gitRepoCache", repoCache))
	r.Use(middleware.WithValue("alertStateManager", alertStateManager))
	r.Use(middleware.WithValue("chartUpdatePullRequests", chartUpdatePullRequests))
	r.Use(middleware.WithValue("gitopsUpdatePullRequests", gitopsUpdatePullRequests))
	r.Use(middleware.WithValue("router", r))
	r.Use(middleware.WithValue("gitUser", gitUser))
	r.Use(middleware.WithValue("gitopsQueue", gitopsQueue))
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
	userRoutes(r, clientHub)
	githubOAuthRoutes(config, dynamicConfig, r)
	gimletdRoutes(r)
	adminKeyAuthRoutes(r)
	installerRoutes(r)

	r.Get("/logout", logout)
	r.Handle("/builtin/infra*", gitServer)
	r.Handle("/builtin/apps*", gitServer)

	r.Post("/hook", hook)

	r.Get("/flags", getFlags)

	filesDir := http.Dir("./web/build")
	fileServer(r, "/", filesDir)
	fileServer(r, "/login", filesDir)
	fileServer(r, "/repositories", filesDir)
	fileServer(r, "/profile", filesDir)
	fileServer(r, "/settings", filesDir)
	fileServer(r, "/repo", filesDir)
	fileServer(r, "/environments", filesDir)
	fileServer(r, "/env", filesDir)

	return r
}

func gimletdRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
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
		r.Get("/api/gitopsManifests/{env}", getGitopsManifests)
	})

	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Use(session.MustAdmin())
		r.Get("/api/user/{login}", getUser)
		r.Post("/api/user", saveUserGimletD)
		r.Post("/api/deleteUser", deleteUser)
		r.Get("/api/users", getUsers)
	})
}

func userRoutes(r *chi.Mux, clientHub *streaming.ClientHub) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(session.SetUser())
		r.Use(session.SetCSRF())
		r.Use(session.MustUser())

		r.Get("/api/agents", agents)
		r.Get("/api/user", user)
		r.Post("/api/saveUser", saveUser)
		r.Get("/api/envs", envs)
		r.Get("/api/fluxEvents", fluxK8sEvents)
		r.Get("/api/podLogs", getPodLogs)
		r.Get("/api/stopPodLogs", stopPodLogs)
		r.Get("/api/deploymentDetails", getDeploymentDetails)
		r.Get("/api/podDetails", getPodDetails)
		r.Post("/api/reconcile", reconcile)
		r.Get("/api/alerts", getAlerts)
		r.Get("/api/gitRepos", gitRepos)
		r.Get("/api/refreshRepos", refreshRepos)
		r.Get("/api/settings", settings)
		r.Get("/api/repo/{owner}/{name}/commits", commits)
		r.Get("/api/repo/{owner}/{name}/commits/{sha}/events", commitEvents)
		r.Get("/api/repo/{owner}/{name}/triggerCommitSync", triggerCommitSync)
		r.Get("/api/gitopsCommits", getGitopsCommits)
		r.Get("/api/repo/{owner}/{name}/branches", branches)
		r.Get("/api/repo/{owner}/{name}/metas", getMetas)
		r.Get("/api/repo/{owner}/{name}/pullRequests", getPullRequests)
		r.Get("/api/chartUpdatePullRequests", getChartUpdatePullRequests)
		r.Get("/api/gitopsUpdatePullRequests", getGitopsUpdatePullRequests)
		r.Get("/api/infraRepoPullRequests", getPullRequestsFromInfraRepos)
		r.Get("/api/repo/{owner}/{name}/envConfigs", envConfigs)
		r.Post("/api/repo/{owner}/{name}/env/{env}/config/{config}", saveEnvConfig)
		r.Post("/api/repo/{owner}/{name}/env/{env}/config/{config}/delete", deleteEnvConfig)
		r.Post("/api/saveFavoriteRepos", saveFavoriteRepos)
		r.Post("/api/saveFavoriteServices", saveFavoriteServices)
		r.Get("/api/defaultDeploymentTemplates", defaultDeploymentTemplates)
		r.Get("/api/repo/{owner}/{name}/env/{env}/config/{config}/deploymentTemplates", deploymentTemplateForApp)
		r.Get(("/api/app"), application)
		r.Post(("/api/saveEnvToDB"), saveEnvToDB)
		r.Post(("/api/spinOutBuiltInEnv"), spinOutBuiltInEnv)
		r.Post(("/api/deleteEnvFromDB"), deleteEnvFromDB)
		r.Post(("/api/environments"), saveInfrastructureComponents)
		r.Post(("/api/bootstrapGitops"), bootstrapGitops)
		r.Post(("/api/env/{env}/seal"), seal)
		r.Get(("/api/env/{env}/stackConfig"), stackConfig)
		r.Post("/api/silenceAlert", silenceAlert)
		r.Post("/api/restartDeployment", restartDeployment)

		r.Get("/ws/", func(w http.ResponseWriter, r *http.Request) {
			streaming.ServeWs(clientHub, w, r)
		})
	})
}

func agentRoutes(r *chi.Mux, agentWSHub *streaming.AgentWSHub) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(session.SetUser())
		r.Use(combinedAuthorizer)

		r.Post("/agent/state", state)
		r.Post("/agent/state/{name}/update", update)
		r.Post("/agent/events", events)
		r.Post("/agent/fluxState", fluxState)
		r.Post("/agent/fluxEvents", sendFluxEvents)
		r.Post("/agent/deploymentDetails", deploymentDetails)
		r.Post("/agent/podDetails", podDetails)
		r.Get("/agent/ws/", func(w http.ResponseWriter, r *http.Request) {
			streaming.ServeAgentWs(agentWSHub, w, r)
		})
	})
	r.Group(func(r chi.Router) { // group with one hour timeout
		r.Use(session.SetUser())
		r.Use(combinedAuthorizer)
		r.Use(middleware.Timeout(60 * time.Minute))

		r.Get("/agent/register", register)
	})
	r.Get("/agent/imagebuild/{imageBuildId}", imageBuild)
}

func githubOAuthRoutes(config *config.Config, dynamicConfig *dynamicconfig.DynamicConfig, r *chi.Mux) {
	if dynamicConfig.IsGithub() {
		dumper := logger.DiscardDumper()
		if dynamicConfig.Github.Debug {
			dumper = logger.StandardDumper()
		}
		loginMiddleware := &github.Config{
			ClientID:     dynamicConfig.Github.ClientID,
			ClientSecret: dynamicConfig.Github.ClientSecret,
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
	} else if dynamicConfig.IsGitlab() {
		loginMiddleware := &gitlab.Config{
			Server:       dynamicConfig.ScmURL(),
			ClientID:     dynamicConfig.Gitlab.ClientID,
			ClientSecret: dynamicConfig.Gitlab.ClientSecret,
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

func adminKeyAuthRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(session.SetUser())
		r.Post("/admin-key-auth", adminKeyAuth)
	})
}

func installerRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(session.SetUser())
		r.Use(session.MustUser())
		r.Get("/settings/created", created)
		r.Get("/settings/installed", installed)
		r.Post("/settings/gitlabInit", gitlabInit)
	})
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
			pathPrefix == "/env" {
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

func combinedAuthorizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// check if a user is authenticated
		_, userSet := ctx.Value("user").(*model.User)
		if userSet {
			next.ServeHTTP(w, r)
			return
		}

		// do agent authentication and authorization
		dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
		agentAuth = jwtauth.New("HS256", []byte(dynamicConfig.JWTSecret), nil)

		verifierFunc := jwtauth.Verifier(agentAuth)
		authenticatorFunc := jwtauth.Authenticator

		verifierFunc(authenticatorFunc(mustAgent(next))).ServeHTTP(w, r)
	})
}
