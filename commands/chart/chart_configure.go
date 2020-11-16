package chart

import (
	"context"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/go-chi/chi"
	"github.com/urfave/cli/v2"
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
	Usage:     "Configures Helm chart values",
	ArgsUsage: "<repo/name>",
	Action:    configure,
}

func configure(c *cli.Context) error {
	port := randomPort()

	workDir, err := ioutil.TempDir(os.TempDir(), "gimlet")
	if err != nil {
		panic(err)
	}
	fmt.Println(workDir)
	writeTempFiles(workDir)
	defer removeTempFiles(workDir)
	r := setupRouter(workDir)
	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: chi.ServerBaseContext(context.TODO(), r)}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	go srv.ListenAndServe()
	fmt.Printf("%v Configure on http://127.0.0.1:%d\n", emoji.WomanTechnologist, port)
	fmt.Printf("%v Close the browser when you are done\n\n", emoji.WomanTechnologist)
	openBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))

	<-ctrlC
	fmt.Printf("%v Generating values..\n\n", emoji.FileFolder)
	srv.Shutdown(context.TODO())

	return nil
}

func removeTempFiles(workDir string) {
	for file, _ := range web {
		os.Remove(filepath.Join(workDir, file))
	}
	os.Remove(workDir)
}

func writeTempFiles(workDir string) {
	for file, content := range web {
		ioutil.WriteFile(filepath.Join(workDir, file), []byte(content), 0666)
	}
}

func setupRouter(workDir string) *chi.Mux {
	r := chi.NewRouter()
	//r.Use(middleware.Logger)

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	filesDir := http.Dir(workDir)
	fileServer(r, "/", filesDir)

	return r
}

func randomPort() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000)+20000
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
