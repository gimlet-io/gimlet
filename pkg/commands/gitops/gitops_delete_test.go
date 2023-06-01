package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
)

func Test_delete(t *testing.T) {
	gitopsRepoPath, err := ioutil.TempDir("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitopsRepoPath)

	repo, _ := git.PlainInit(gitopsRepoPath, false)

	t.Run("gimlet gitops delete", func(t *testing.T) {
		t.Run("Should validate path exist", func(t *testing.T) {
			args := strings.Split("gimlet gitops delete", " ")
			args = append(args, "--gitops-repo-path", "does-not-exist")
			args = append(args, "--env", "staging")
			args = append(args, "--app", "my-app")
			err := commands.Run(&Command, args)
			assert.Contains(t, err.Error(), "is not a git repo")
		})

		args := strings.Split("gimlet gitops delete", " ")
		args = append(args, "--gitops-repo-path", gitopsRepoPath)
		args = append(args, "--env", "staging")
		args = append(args, "--app", "my-app")

		t.Run("Should stage and commit a folder", func(t *testing.T) {
			err := os.MkdirAll(filepath.Join(gitopsRepoPath, "staging", "my-app"), os.ModePerm)
			assert.NoError(t, err)
			err = ioutil.WriteFile(filepath.Join(gitopsRepoPath, "staging", "my-app", "dummy"), []byte(""), os.ModePerm)
			assert.NoError(t, err)
			err = nativeGit.StageFolder(repo, "staging")
			assert.NoError(t, err)
			_, err = nativeGit.Commit(repo, "")
			assert.NoError(t, err)
		})

		t.Run("Should delete path with a commit", func(t *testing.T) {
			err := commands.Run(&Command, args)
			assert.NoError(t, err)

			head, err := repo.Head()
			assert.NoError(t, err)
			lastCommit, err := repo.CommitObject(head.Hash())
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("[Gimlet CLI delete] %s/%s %s", "staging", "my-app", ""), lastCommit.Message)
		})
	})
}
