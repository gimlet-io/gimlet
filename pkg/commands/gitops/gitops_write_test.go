package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/go-git/go-git/v5"
)

func Test_write(t *testing.T) {
	gitopsRepoPath, err := ioutil.TempDir("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitopsRepoPath)

	repo, _ := git.PlainInit(gitopsRepoPath, false)

	const env = "staging"
	const app = "my-app"

	args := strings.Split("gimlet gitops write", " ")
	args = append(args, "--gitops-repo-path", gitopsRepoPath)
	args = append(args, "--env", env)
	args = append(args, "--app", app)

	t.Run("Should write a file", func(t *testing.T) {
		fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(fileToWrite.Name())

		ioutil.WriteFile(fileToWrite.Name(), []byte("dummyContent"), commands.File_RW_RW_R)
		args = append(args, "-f", fileToWrite.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}

		head, err := repo.Head()
		if err != nil {
			t.Fatal(err)
		}
		lastCommit, err := repo.CommitObject(head.Hash())
		if err != nil {
			t.Fatal(err)
		}
		if lastCommit.Message != fmt.Sprintf("[Gimlet] %s/%s %s", env, app, "") {
			t.Errorf("Expected commit message to be '[Gimlet] %s/%s %s', got '%s'", env, app, "", lastCommit.Message)
		}
	})

	t.Run("Should write a folder", func(t *testing.T) {
		dirToWrite, err := ioutil.TempDir("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dirToWrite)

		ioutil.WriteFile(filepath.Join(dirToWrite, "dummy"), []byte("dummyContent"), commands.File_RW_RW_R)
		args = append(args, "-f", dirToWrite)

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}

		head, err := repo.Head()
		if err != nil {
			t.Fatal(err)
		}
		lastCommit, err := repo.CommitObject(head.Hash())
		if err != nil {
			t.Fatal(err)
		}
		if lastCommit.Message != fmt.Sprintf("[Gimlet] %s/%s %s", env, app, "") {
			t.Errorf("Expected commit message to be '[Gimlet] %s/%s %s', got '%s'", env, app, "", lastCommit.Message)
		}
	})

	t.Run("Should split a templated Helm manifest to multiple files", func(t *testing.T) {
		deleteArgs := strings.Split("gimlet gitops delete", " ")
		deleteArgs = append(deleteArgs, "--gitops-repo-path", gitopsRepoPath)
		deleteArgs = append(deleteArgs, "--env", env)
		deleteArgs = append(deleteArgs, "--app", app)
		err = commands.Run(&Command, deleteArgs)
		if err != nil {
			t.Fatal(err)
		}

		fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(fileToWrite.Name())

		ioutil.WriteFile(fileToWrite.Name(), []byte(manifestStr), commands.File_RW_RW_R)
		args = append(args, "-f", fileToWrite.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}

		scoreBoard := map[string]bool{
			"service.yaml":    false,
			"deployment.yaml": false,
		}

		dir, err := ioutil.ReadDir(filepath.Join(gitopsRepoPath, env, app))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range dir {
			if f.IsDir() {
				continue
			}
			_, err := ioutil.ReadFile(filepath.Join(gitopsRepoPath, env, app, f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			scoreBoard[f.Name()] = true
		}

		for k, v := range scoreBoard {
			if !v {
				t.Errorf("Couldn't find file %s", k)
			}
		}
	})
}

const manifestStr = `
---
# Source: onechart/templates/service.yaml
apiVersion: v1
kind: Service
---
# Source: onechart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
`
