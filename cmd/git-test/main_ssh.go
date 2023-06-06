package main

import (
	"log"

	"github.com/sosedoff/gitkit"
)

// User-defined key lookup function. You can make a call to a database or
// some sort of cache storage (redis/memcached) to speed things up.
// Content is a string containing ssh public key of a user.
func lookupKey(content string) (*gitkit.PublicKey, error) {
	return &gitkit.PublicKey{Id: "ssh-rsa AAAAB3NzaC1y....."}, nil
}

func main_ssh() {
	// In the example below you need to specify a full path to a directory that
	// contains all git repositories, and also a directory that has a gitkit specific
	// ssh private and public key pair that used to run ssh server.
	server := gitkit.NewSSH(gitkit.Config{
		Dir:    "/home/laszlo/projects/gimlet/git-server-root",
		KeyDir: "/home/laszlo/projects/gimlet/git-server-keys",
	})

	// User-defined key lookup function. All requests will be rejected if this function
	// is not provider. SSH server only accepts key-based authentication.
	server.PublicKeyLookupFunc = lookupKey

	// Specify host and port to run the server on.
	err := server.ListenAndServe(":2222")
	if err != nil {
		log.Fatal(err)
	}
}
