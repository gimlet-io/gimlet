package gitops

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var gitopsBootstrapCmd = cli.Command{
	Name:      "bootstrap",
	Usage:     "Bootstraps the gitops controller for an environment",
	UsageText: `gimlet gitops bootstrap \
     --env staging \
     --gitops-repo-url https://github.com/<user>/<repo>.git`,
	Action:    bootstrap,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "env",
			Usage: "environment to write to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-url",
			Usage: "URL of the gitops repo (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo",
		},
	},
}

func bootstrap(c *cli.Context) error {
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

	empty, err := nothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are staged changes in the gitops repo. Commit them first then try again")
	}

	env := c.String("env")
	//gitopsRepoUrl := c.String("gitops-repo-url")



	empty, err = nothingToCommit(repo)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gitMessage := fmt.Sprintf("[Gimlet CLI bootstrap] %s", env)
	return commit(repo, gitMessage)
}
