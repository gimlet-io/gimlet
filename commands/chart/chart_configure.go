package chart

import (
	"context"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/go-chi/chi"
	"github.com/urfave/cli/v2"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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

	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	srv := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: chi.ServerBaseContext(context.TODO(), r)}

	//https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=&cad=rja&uact=8&ved=2ahUKEwi3--283P_sAhXHAhAIHZj8CNEQFjAAegQIARAC&url=https%3A%2F%2Fgithub.com%2Fgo-chi%2Fvalve&usg=AOvVaw3Z32QPj1RXr6hqQfjXMT2C
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	go srv.ListenAndServe()
	fmt.Printf("%v Configure on http://127.0.0.1:%d\n", emoji.WomanTechnologist, port)
	fmt.Printf("%v Close the browser when you are done\n\n", emoji.WomanTechnologist)
	openBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))

	<- ctrlC
	fmt.Printf("%v Generating values..\n\n", emoji.FileFolder)
	srv.Shutdown(context.TODO())

	return nil
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
