package gitops

import (
	"fmt"
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_write(t *testing.T) {
	gitopsRepoPath, err := ioutil.TempDir("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitopsRepoPath)

	repo, _ := git.PlainInit(gitopsRepoPath, false)

	g := goblin.Goblin(t)

	const env = "staging"
	const app = "my-app"

	args := strings.Split("gimlet gitops write", " ")
	args = append(args, "--gitops-repo-path", gitopsRepoPath)
	args = append(args, "--env", env)
	args = append(args, "--app", app)

	g.Describe("gimlet gitops write", func() {
		g.It("Should write a file", func() {
			fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(fileToWrite.Name())

			ioutil.WriteFile(fileToWrite.Name(), []byte("dummyContent"), commands.File_RW_RW_R)
			args = append(args, "-f", fileToWrite.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue(err)
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue(err)
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet] %s/%s %s", env, app, ""))
		})

		g.It("Should write a folder", func() {
			dirToWrite, err := ioutil.TempDir("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dirToWrite)

			ioutil.WriteFile(filepath.Join(dirToWrite, "dummy"), []byte("dummyContent"), commands.File_RW_RW_R)
			args = append(args, "-f", dirToWrite)

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue(err)
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue(err)
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet] %s/%s %s", env, app, ""))
		})

		g.It("Should split a templated Helm manifest to multiple files", func() {
			deleteArgs := strings.Split("gimlet gitops delete", " ")
			deleteArgs = append(deleteArgs, "--gitops-repo-path", gitopsRepoPath)
			deleteArgs = append(deleteArgs, "--env", env)
			deleteArgs = append(deleteArgs, "--app", app)
			err = commands.Run(&Command, deleteArgs)
			g.Assert(err == nil).IsTrue(err)

			fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(fileToWrite.Name())

			ioutil.WriteFile(fileToWrite.Name(), []byte(manifestStr), commands.File_RW_RW_R)
			args = append(args, "-f", fileToWrite.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			scoreBoard := map[string]bool{
				"service.yaml":    false,
				"deployment.yaml": false,
			}

			dir, err := ioutil.ReadDir(filepath.Join(gitopsRepoPath, env, app))
			g.Assert(err == nil).IsTrue(err)
			for _, f := range dir {
				if f.IsDir() {
					continue
				}
				_, err := ioutil.ReadFile(filepath.Join(gitopsRepoPath, env, app, f.Name()))
				g.Assert(err == nil).IsTrue(err)
				scoreBoard[f.Name()] = true
			}

			for k, v := range scoreBoard {
				g.Assert(v).IsTrue("Couldn't find file ", k)
			}

		})
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
