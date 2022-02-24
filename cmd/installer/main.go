package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.WithValue("app", &app{}))

	r.Post("/saveAppCredentials", saveAppCredentials)
	r.HandleFunc("/*", serveTemplate)

	http.ListenAndServe(":3000", r)
}

type app struct {
	id           string
	clientId     string
	clientSecret string
	pem          string
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "index.html"
	} else if !strings.HasSuffix(path, ".html") {
		path = path + ".html"
	}
	fp := filepath.Join("web", path)

	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		fmt.Println(err.Error())
		return
	}

	ctx := r.Context()
	app := ctx.Value("app").(*app)

	err = tmpl.Execute(w, map[string]string{
		"appId":        app.id,
		"clientId":     app.clientId,
		"clientSecret": app.clientSecret,
		"pem":          app.pem,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func saveAppCredentials(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	ctx := r.Context()
	app := ctx.Value("app").(*app)

	formValues := r.PostForm
	app.id = formValues.Get("appId")
	app.clientId = formValues.Get("clientId")
	app.clientSecret = formValues.Get("clientSecret")
	app.pem = formValues.Get("pem")

	w.WriteHeader(http.StatusOK)
}
