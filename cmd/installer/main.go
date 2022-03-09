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

	r.Use(middleware.WithValue("data", &data{}))

	r.Post("/saveAppCredentials", saveAppCredentials)
	r.Post("/bootstrap", bootstrap)
	r.HandleFunc("/*", serveTemplate)

	http.ListenAndServe(":3333", r)
}

type data struct {
	id           string
	clientId     string
	clientSecret string
	pem          string
	org          string
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
	data := ctx.Value("data").(*data)

	err = tmpl.Execute(w, map[string]string{
		"appId":        data.id,
		"clientId":     data.clientId,
		"clientSecret": data.clientSecret,
		"pem":          data.pem,
		"org":          data.org,
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
	data := ctx.Value("data").(*data)

	formValues := r.PostForm
	data.id = formValues.Get("appId")
	data.clientId = formValues.Get("clientId")
	data.clientSecret = formValues.Get("clientSecret")
	data.pem = formValues.Get("pem")

	

	w.WriteHeader(http.StatusOK)
}

func bootstrap(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	formValues := r.PostForm
	fmt.Println(formValues)
}
