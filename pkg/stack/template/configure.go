package template

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/chart/ws"
	"github.com/gimlet-io/gimlet-cli/pkg/stack/template/web"
	"github.com/gimlet-io/gimlet-cli/pkg/version"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

type Component struct {
	Name        string `json:"name,omitempty" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description"`
	Category    string `json:"category,omitempty" yaml:"category"`
	Variable    string `json:"variable,omitempty" yaml:"variable"`
	Logo        string `json:"logo,omitempty" yaml:"logo"`
	OnePager    string `json:"onePager,omitempty" yaml:"onePager"`
	Schema      string `json:"schema,omitempty" yaml:"schema"`
	UISchema    string `json:"uiSchema,omitempty" yaml:"uiSchema"`
}

type StackDefinition struct {
	Name        string        `json:"name,omitempty" yaml:"name"`
	Description string        `json:"description,omitempty" yaml:"description"`
	Intro       string        `json:"intro,omitempty" yaml:"intro"`
	Categories  []interface{} `json:"categories" yaml:"categories"`
	Components  []*Component  `json:"components,omitempty" yaml:"components"`
	ChangLog    string        `json:"changeLog,omitempty" yaml:"changeLog"`
	Message     string        `json:"message,omitempty" yaml:"message"`
}

func StackDefinitionFromRepo(repoUrl string) (string, error) {
	stackTemplates, err := cloneStackFromRepo(repoUrl)
	if err != nil {
		return "", err
	}

	return stackTemplates["stack-definition.yaml"], nil
}

var values map[string]interface{}
var written bool

func Configure(stackDefinition StackDefinition, existingStackConfig StackConfig) (StackConfig, bool, error) {
	stackDefinitionJson, err := json.Marshal(stackDefinition)
	if err != nil {
		panic(err)
	}

	if existingStackConfig.Config == nil {
		existingStackConfig.Config = map[string]interface{}{}
	}

	values = existingStackConfig.Config

	stackJson, err := json.Marshal(existingStackConfig.Config)
	if err != nil {
		panic(err)
	}

	port := randomPort()

	workDir, err := ioutil.TempDir(os.TempDir(), "gimlet")
	if err != nil {
		panic(err)
	}
	writeTempFiles(workDir, string(stackDefinitionJson), string(stackJson))
	defer removeTempFiles(workDir)
	browserClosed := make(chan int, 1)
	r := setupRouter(workDir, browserClosed)
	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	go srv.ListenAndServe()
	fmt.Fprintf(os.Stderr, "%v Configure on http://127.0.0.1:%d\n", emoji.WomanTechnologist, port)
	fmt.Fprintf(os.Stderr, "%v Close the browser when you are done\n", emoji.WomanTechnologist)
	err = openBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}

	select {
	case <-ctrlC:
	case <-browserClosed:
	}

	fmt.Fprintf(os.Stderr, "%v Generating values..\n\n", emoji.FileFolder)
	srv.Shutdown(context.TODO())

	existingStackConfig.Config = values

	return existingStackConfig, written, nil
}

func randomPort() int {
	if version.String() == "idea" {
		return 28000
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000) + 20000
}

func writeTempFiles(workDir string, stackDefinition string, stackJson string) {
	ioutil.WriteFile(filepath.Join(workDir, "stack-definition.json"), []byte(stackDefinition), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "stack.json"), []byte(stackJson), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "bundle.js"), web.BundleJs, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "bundle.js.LICENSE.txt"), web.LicenseTxt, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "index.html"), web.IndexHtml, 0666)
}

func removeTempFiles(workDir string) {
	os.Remove(workDir)
}

func setupRouter(workDir string, browserClosed chan int) *chi.Mux {
	r := chi.NewRouter()
	if version.String() == "idea" {
		r.Use(middleware.Logger)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:28000", "http://127.0.0.1:28000"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(browserClosed, w, r)
	})

	r.Post("/saveValues", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&values)
		written = true
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	})

	filesDir := http.Dir(workDir)
	fileServer(r, "/", filesDir)

	return r
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
