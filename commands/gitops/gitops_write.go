package gitops

import (
	"fmt"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/gimlet-io/gimletd/githelper"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var gitopsWriteCmd = cli.Command{
	Name:  "write",
	Usage: "Writes app manifests to a gitops environment",
	UsageText: `gimlet gitops write -f my-app.yaml \
     --env staging \
     --app my-app \
     -m "Releasing Bugfix 345"`,
	Action: write,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "manifest file, folder or \"-\" for stdin to write (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to write to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "app",
			Usage:    "name of the application that you configure (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo, default: current dir",
		},
		&cli.StringFlag{
			Name:    "message",
			Aliases: []string{"m"},
			Usage:   "gitops commit message",
		},
	},
}

func write(c *cli.Context) error {
	gitopsRepoPath := c.String("gitops-repo-path")
	if gitopsRepoPath == "" {
		gitopsRepoPath, _ = os.Getwd()
	}
	gitopsRepoPath, err := filepath.Abs(gitopsRepoPath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path %s", err)
	}

	repo, err := git.PlainOpen(gitopsRepoPath)
	if err == git.ErrRepositoryNotExists {
		return fmt.Errorf("%s is not a git repository", gitopsRepoPath)
	}

	env := c.String("env")
	app := c.String("app")
	file := c.String("file")
	message := c.String("message")

	files, err := commands.InputFiles(file)
	if err != nil {
		return fmt.Errorf("cannot read input files %s", err)
	}
	files = dx.SplitHelmOutput(files)

	_, err = githelper.CommitFilesToGit(repo, files, env, app, message, "")
	return err
}
