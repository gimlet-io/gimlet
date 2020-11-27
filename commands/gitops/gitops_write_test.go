package gitops

import (
	"fmt"
	"github.com/franela/goblin"
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
		g.It("Should validate path exist", func() {
			err = run(args)
			g.Assert(strings.Contains(err.Error(), "Run `gimlet gitops write --help` for usage")).IsTrue()
		})

		g.It("Should write a file", func() {
			fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(fileToWrite.Name())

			ioutil.WriteFile(fileToWrite.Name(), []byte("dummyContent"), file_RW_RW_R)
			args = append(args, "-f", fileToWrite.Name())

			err = run(args)
			g.Assert(err == nil).IsTrue()

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue()
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue()
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet CLI write] %s/%s %s", env, app, ""))
		})

		g.It("Should write a folder", func() {
			dirToWrite, err := ioutil.TempDir("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dirToWrite)

			ioutil.WriteFile(filepath.Join(dirToWrite, "dummy"), []byte("dummyContent"), file_RW_RW_R)
			args = append(args, "-f", dirToWrite)

			err = run(args)
			g.Assert(err == nil).IsTrue(err)

			head, err := repo.Head()
			g.Assert(err == nil).IsTrue()
			lastCommit, err := repo.CommitObject(head.Hash())
			g.Assert(err == nil).IsTrue()
			g.Assert(lastCommit.Message).Equal(fmt.Sprintf("[Gimlet CLI write] %s/%s %s", env, app, ""))
		})
	})

}
