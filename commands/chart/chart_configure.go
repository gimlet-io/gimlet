package chart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/commands/chart/ws"
	"github.com/gimlet-io/gimlet-cli/version"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var chartConfigureCmd = cli.Command{
	Name:      "configure",
	Usage:     "configure Helm chart values",
	ArgsUsage: "<repo/name>",
	Action:    configure,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "edit existing values file",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output values file",
		},
	},
}

var values map[string]interface{}

func configure(c *cli.Context) error {
	repoArg := c.Args().First()
	if repoArg == "" {
		cli.ShowCommandHelp(c, "configure")
		os.Exit(1)
	}

	chartLoader := action.NewShow(action.ShowChart)
	var settings = helmCLI.New()
	chartPath, err := chartLoader.ChartPathOptions.LocateChart(repoArg, settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Could not load %s Helm chart\n", emoji.CrossMark, err.Error())
		os.Exit(1)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Could not load %s Helm chart\n", emoji.CrossMark, err.Error())
		os.Exit(1)
	}

	schema := string(chart.Schema)
	var helmUISchema string
	for _, r := range chart.Raw {
		if "helm-ui.json" == r.Name {
			helmUISchema = string(r.Data)
		}
	}

	if schema == "" {
		fmt.Fprintf(os.Stderr, "%s Chart doesn't have a values.schema.json with the Helm schema defined\n", emoji.CrossMark)
		os.Exit(1)
	}

	if helmUISchema == "" {
		fmt.Fprintf(os.Stderr, "%s Chart doesn't have a helm-ui.json with the Helm UI schema defined\n", emoji.CrossMark)
		os.Exit(1)
	}

	existingValuesPath := c.String("file")
	existingValuesJson := []byte("{}")
	if existingValuesPath != "" {
		yamlString, err := ioutil.ReadFile(existingValuesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Cannot read values file\n", emoji.CrossMark)
			os.Exit(1)
		}

		var parsedYaml map[string]interface{}
		err = yaml.Unmarshal(yamlString, &parsedYaml)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Cannot parse values file\n", emoji.CrossMark)
			os.Exit(1)
		}

		existingValuesJson, err = json.Marshal(parsedYaml)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Cannot serialize values file\n", emoji.CrossMark)
			os.Exit(1)
		}
	}

	port := randomPort()

	workDir, err := ioutil.TempDir(os.TempDir(), "gimlet")
	if err != nil {
		panic(err)
	}
	writeTempFiles(workDir, schema, helmUISchema, string(existingValuesJson))
	defer removeTempFiles(workDir)
	browserClosed := make(chan int, 1)
	r := setupRouter(workDir, browserClosed)
	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: chi.ServerBaseContext(context.TODO(), r)}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	go srv.ListenAndServe()
	fmt.Fprintf(os.Stderr, "%v Configure on http://127.0.0.1:%d\n", emoji.WomanTechnologist, port)
	fmt.Fprintf(os.Stderr, "%v Close the browser when you are done\n", emoji.WomanTechnologist)
	openBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))

	select {
	case <-ctrlC:
	case <-browserClosed:
	}

	fmt.Fprintf(os.Stderr, "%v Generating values..\n\n", emoji.FileFolder)
	srv.Shutdown(context.TODO())

	yamlString := bytes.NewBufferString("")
	e := yaml.NewEncoder(yamlString)
	e.SetIndent(2)
	e.Encode(values)

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, yamlString.Bytes(), 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Cannot write values file\n", emoji.CrossMark)
			os.Exit(1)
		}
	} else {
		fmt.Println("---")
		fmt.Println(yamlString.String())
	}

	return nil
}

func removeTempFiles(workDir string) {
	for file, _ := range web {
		os.Remove(filepath.Join(workDir, file))
	}
	os.Remove(workDir)
}

func writeTempFiles(workDir string, schema string, helmUISchema string, existingValues string) {
	ioutil.WriteFile(filepath.Join(workDir, "values.schema.json"), []byte(schema), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "helm-ui.json"), []byte(helmUISchema), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "values.json"), []byte(existingValues), 0666)

	for file, content := range web {
		ioutil.WriteFile(filepath.Join(workDir, file), []byte(content), 0666)
	}
}

func setupRouter(workDir string, browserClosed chan int) *chi.Mux {
	r := chi.NewRouter()
	if version.String() == "idea" {
		//r.Use(middleware.Logger)
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
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	})

	filesDir := http.Dir(workDir)
	fileServer(r, "/", filesDir)

	return r
}

func randomPort() int {
	if version.String() == "idea" {
		return 28000
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000) + 20000
}

func openBrowser(url string) {
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
	if err != nil {
		log.Fatal(err)
	}
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
