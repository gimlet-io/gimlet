package gitops

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"path/filepath"
	"time"
)

func nothingToCommit(repo *git.Repository) (bool, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, err
	}

	return status.IsClean(), nil
}

func commit(repo *git.Repository, message string, env string, app string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Gimlet CLI",
			Email: "cli@gimlet.io",
			When:  time.Now(),
		},
	})

	return err
}

func delDir(repo *git.Repository, path string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	files, err := worktree.Filesystem.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			delDir(repo, file.Name())
		}

		_, err = worktree.Remove(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
	}

	_, err = worktree.Remove(path)

	return err
}
