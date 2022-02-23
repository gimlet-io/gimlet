package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/saveAppCredentials", saveAppCredentials)
	r.HandleFunc("/*", serveTemplate)

	http.ListenAndServe(":3000", r)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "index.html"
	}
	fp := filepath.Join("web", path)

	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		fmt.Println(err.Error())
		return
	}
	err = tmpl.Execute(w, map[string]string{})
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
	fmt.Println(r.PostForm)

	w.WriteHeader(http.StatusOK)
}
