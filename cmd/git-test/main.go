package main

import (
	"log"
	"net/http"

	"github.com/sosedoff/gitkit"
)

func main() {
	// Configure git hooks
	hooks := &gitkit.HookScripts{}

	// Configure git service
	service := gitkit.New(gitkit.Config{
		Dir:        "/home/laszlo/projects/gimlet/git-server-root",
		AutoCreate: true,
		AutoHooks:  true,
		Hooks:      hooks,
		Auth:       true,
	})

	// Here's the user-defined authentication function.
	// If return value is false or error is set, user's request will be rejected.
	// You can hook up your database/redis/cache for authentication purposes.
	service.AuthFunc = func(cred gitkit.Credential, req *gitkit.Request) (bool, error) {
		log.Println("user auth request for repo:", cred.Username, cred.Password, req.RepoName)
		return cred.Username == "testuser" && cred.Password == "49bec54a", nil
	}

	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := service.Setup(); err != nil {
		log.Fatal(err)
	}

	http.Handle("/", service)

	// Start HTTP server
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal(err)
	}
}
