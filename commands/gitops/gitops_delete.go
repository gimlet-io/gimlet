package gitops

import (
	"fmt"
	"github.com/gimlet-io/gimletd/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var gitopsDeleteCmd = cli.Command{
	Name:  "delete",
	Usage: "Deletes app manifests from an environment",
	UsageText: `gimlet gitops delete \
     --env staging \
     --app my-app-review-bugfix-345
     -m "Dropping preview environment for Bugfix 345"`,
	Action: delete,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to write to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name: "app",
			Usage: "name of the application that you configure (	mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo",
		},
		&cli.StringFlag{
			Name:    "message",
			Aliases: []string{"m"},
			Usage:   "gitops commit message",
		},
	},
}

func delete(c *cli.Context) error {
	gitopsRepoPath := c.String("gitops-repo-path")
	if gitopsRepoPath == "" {
		gitopsRepoPath, _ = os.Getwd()
	}
	gitopsRepoPath, err := filepath.Abs(gitopsRepoPath)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(gitopsRepoPath)
	if err == git.ErrRepositoryNotExists {
		return fmt.Errorf("%s is not a git repo\n", gitopsRepoPath)
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are staged changes in the gitops repo. Commit them first then try again")
	}

	env := c.String("env")
	app := c.String("app")
	message := c.String("message")

	err = nativeGit.DelDir(repo, filepath.Join(env, app))
	if err != nil {
		return err
	}

	empty, err = nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gitMessage := fmt.Sprintf("[Gimlet CLI delete] %s/%s %s", env, app, message)
	_, err = nativeGit.Commit(repo, gitMessage)
	return err
}
