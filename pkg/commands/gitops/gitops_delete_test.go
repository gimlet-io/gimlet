package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
)

func Test_delete(t *testing.T) {
	gitopsRepoPath, err := ioutil.TempDir("", "gimlet-cli-test")
	if err != nil {
		t.Fatalf("Error creating temp directory: %s", err)
	}
	defer os.RemoveAll(gitopsRepoPath)

	repo, _ := git.PlainInit(gitopsRepoPath, false)

	const env = "staging"
	const app = "my-app"

	t.Run("Should validate path exist", func(t *testing.T) {
		args := strings.Split("gimlet gitops delete", " ")
		args = append(args, "--gitops-repo-path", "does-not-exist")
		args = append(args, "--env", env)
		args = append(args, "--app", app)
		if err := commands.Run(&Command, args); err == nil || !strings.Contains(err.Error(), "is not a git repo") {
			t.Fatalf("Expected error 'is not a git repo', got: %v", err)
		}
	})

	args := strings.Split("gimlet gitops delete", " ")
	args = append(args, "--gitops-repo-path", gitopsRepoPath)
	args = append(args, "--env", env)
	args = append(args, "--app", app)

	t.Run("Should stage and commit a folder", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(gitopsRepoPath, env, app), commands.Dir_RWX_RX_R); err != nil {
			t.Fatalf("Error creating directory: %s", err)
		}
		if err := ioutil.WriteFile(filepath.Join(gitopsRepoPath, env, app, "dummy"), []byte(""), commands.File_RW_RW_R); err != nil {
			t.Fatalf("Error writing file: %s", err)
		}
		if err := nativeGit.StageFolder(repo, env); err != nil {
			t.Fatalf("Error staging folder: %s", err)
		}
		if _, err := nativeGit.Commit(repo, ""); err != nil {
			t.Fatalf("Error committing: %s", err)
		}
	})

	t.Run("Should delete path with a commit", func(t *testing.T) {
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error running command: %s", err)
		}

		head, err := repo.Head()
		if err != nil {
			t.Fatalf("Error getting HEAD: %s", err)
		}
		lastCommit, err := repo.CommitObject(head.Hash())
		if err != nil {
			t.Fatalf("Error getting last commit: %s", err)
		}
		expectedMessage := fmt.Sprintf("[Gimlet CLI delete] %s/%s %s", env, app, "")
		if lastCommit.Message != expectedMessage {
			t.Fatalf("Expected commit message '%s', got '%s'", expectedMessage, lastCommit.Message)
		}
	})
}
