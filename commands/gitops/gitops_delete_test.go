package gitops

import (
	"fmt"
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/gimlet-io/gimletd/githelper"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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



	g.Describe("gimlet gitops delete", func() {
		g.It("Should validate path exist", func() {
			args := strings.Split("gimlet gitops delete", " ")
			args = append(args, "--gitops-repo-path", "does-not-exist")
			args = append(args, "--env", env)
			args = append(args, "--app", app)
			err = commands.Run(&Command, args)
			g.Assert(strings.Contains(err.Error(), "is not a git repo")).IsTrue()
		})

		args := strings.Split("gimlet gitops delete", " ")
		args = append(args, "--gitops-repo-path", gitopsRepoPath)
		args = append(args, "--env", env)
		args = append(args, "--app", app)

		g.It("Should stage and commit a folder", func() {
			err = os.MkdirAll(filepath.Join(gitopsRepoPath, env, app), commands.Dir_RWX_RX_R)
			ioutil.WriteFile(filepath.Join(gitopsRepoPath, env, app, "dummy"), []byte(""), commands.File_RW_RW_R)
			err = githelper.StageFolder(repo, env)
			g.Assert(err == nil).IsTrue()
			_, err = githelper.Commit(repo, "")
			g.Assert(err == nil).IsTrue()
		})

		g.It("Should delete path with a commit", func() {
			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue()

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue()
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue()
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet CLI delete] %s/%s %s", env, app, ""))
		})

	})
}
