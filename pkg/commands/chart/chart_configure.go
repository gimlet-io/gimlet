package chart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/chart/ws"
	"github.com/gimlet-io/gimlet-cli/pkg/version"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
)

var chartConfigureCmd = cli.Command{
	Name:      "configure",
	Usage:     "Configures Helm chart values",
	UsageText: `gimlet chart configure onechart/onechart > values.yaml`,
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
		&cli.StringFlag{
			Name:    "schema",
			Aliases: []string{"s"},
			Usage:   "schema file to render, made for schema development",
		},
		&cli.StringFlag{
			Name:    "ui-schema",
			Aliases: []string{"u"},
			Usage:   "ui schema file to render, made for schema development",
		},
	},
}

var values map[string]interface{}

func configure(c *cli.Context) error {
	repoArg := c.Args().First()
	if repoArg == "" {
		return fmt.Errorf("chart is mandatory. Run `gimlet chart configure --help` for usage")
	}

	existingValuesPath := c.String("file")
	existingValuesJson := []byte("{}")
	if existingValuesPath != "" {
		yamlString, err := ioutil.ReadFile(existingValuesPath)
		if err != nil {
			return fmt.Errorf("cannot read values file")
		}

		var parsedYaml map[string]interface{}
		err = yaml.Unmarshal(yamlString, &parsedYaml)
		if err != nil {
			return fmt.Errorf("cannot parse values")
		}

		existingValuesJson, err = json.Marshal(parsedYaml)
		if err != nil {
			return fmt.Errorf("cannot serialize values")
		}
	}

	var debugSchema, debugUISchema string
	if c.String("schema") != "" {
		debugSchemaBytes, err := ioutil.ReadFile(c.String("schema"))
		if err != nil {
			return fmt.Errorf("cannot read debugSchema file")
		}
		debugSchema = string(debugSchemaBytes)
	}
	if c.String("ui-schema") != "" {
		debugUISchemaBytes, err := ioutil.ReadFile(c.String("ui-schema"))
		if err != nil {
			return fmt.Errorf("cannot read debugUISchema file")
		}
		debugUISchema = string(debugUISchemaBytes)
	}

	yamlBytes, err := ConfigureChart(
		repoArg,
		"",
		"",
		existingValuesJson,
		debugSchema,
		debugUISchema,
	)
	if err != nil {
		return err
	}

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, yamlBytes, 0666)
		if err != nil {
			return fmt.Errorf("cannot write values file %s", err)
		}
	} else {
		fmt.Println("---")
		fmt.Println(string(yamlBytes))
	}

	return nil
}

func ConfigureChart(
	chartName string,
	chartRepository string,
	chartVersion string,
	existingValuesJson []byte,
	debugSchema string,
	debugUISchema string,
) ([]byte, error) {
	var schema, helmUISchema string

	if debugSchema == "" && debugUISchema == "" {
		chartLoader := action.NewShow(action.ShowChart)
		var settings = helmCLI.New()

		chartLoader.ChartPathOptions.RepoURL = chartRepository
		chartLoader.ChartPathOptions.Version = chartVersion
		chartPath, err := chartLoader.ChartPathOptions.LocateChart(chartName, settings)
		if err != nil {
			return nil, fmt.Errorf("could not load %s Helm chart", err.Error())
		}

		chart, err := loader.Load(chartPath)
		if err != nil {
			return nil, fmt.Errorf("could not load %s Helm chart", err.Error())
		}

		schema = string(chart.Schema)
		for _, r := range chart.Raw {
			if "helm-ui.json" == r.Name {
				helmUISchema = string(r.Data)
			}
		}
	}

	if debugSchema != "" {
		schema = debugSchema
	}
	if debugUISchema != "" {
		helmUISchema = debugUISchema
	}

	if schema == "" {
		return nil, fmt.Errorf("chart doesn't have a values.schema.json with the Helm schema defined")
	}
	if helmUISchema == "" {
		return nil, fmt.Errorf("chart doesn't have a helm-ui.json with the Helm UI schema defined")
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
	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}

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
	return yamlString.Bytes(), nil
}

func removeTempFiles(workDir string) {
	os.Remove(workDir)
}

func writeTempFiles(workDir string, schema string, helmUISchema string, existingValues string) {
	ioutil.WriteFile(filepath.Join(workDir, "values.schema.json"), []byte(schema), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "helm-ui.json"), []byte(helmUISchema), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "values.json"), []byte(existingValues), 0666)
	ioutil.WriteFile(filepath.Join(workDir, "bundle.js"), bundleJs, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "bundle.js.LICENSE.txt"), licenseTxt, 0666)
	ioutil.WriteFile(filepath.Join(workDir, "index.html"), indexHtml, 0666)
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
