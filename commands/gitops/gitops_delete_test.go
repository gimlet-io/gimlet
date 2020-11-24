package gitops

import (
	"fmt"
	"github.com/franela/goblin"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(args []string) error {
	app := &cli.App{
		Name:                 "gimlet",
		Version:              "test",
		Usage:                "for an open-source GitOps workflow",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&Command,
		},
	}
	return app.Run(args)
}

func Test_delete(t *testing.T) {
	gitopsRepoPath, err := ioutil.TempDir("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitopsRepoPath)

	repo, _ := git.PlainInit(gitopsRepoPath, false)

	g := goblin.Goblin(t)

	const env = "staging"
	const app = "my-app"

	args := strings.Split("gimlet gitops delete", " ")
	args = append(args, "--gitops-repo-path", gitopsRepoPath)
	args = append(args, "--env", env)
	args = append(args, "--app", app)

	g.Describe("gimlet gitops delete", func() {
		g.It("Should validate path exist", func() {
			err = run(args)
			g.Assert(strings.Contains(err.Error(), "no such file or directory")).IsTrue()
		})

		g.It("Should stage and commit a folder", func() {
			err = os.MkdirAll(filepath.Join(gitopsRepoPath, env, app), dir_RWX_RX_R)
			ioutil.WriteFile(filepath.Join(gitopsRepoPath, env, app, "dummy"), []byte(""), file_RW_RW_R)
			err = stageFolder(repo, env)
			g.Assert(err == nil).IsTrue()
			err = commit(repo, "", env, app)
			g.Assert(err == nil).IsTrue()
		})

		g.It("Should delete path with a commit", func() {
			err = run(args)
			g.Assert(err == nil).IsTrue()

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue()
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue()
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet CLI delete] %s/%s %s", env, app, ""))
		})

	})
}
