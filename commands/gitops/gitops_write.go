package gitops

import (
	"fmt"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	empty, err := nothingToCommit(repo)
	if err != nil {
		return fmt.Errorf("cannot get git state %s", err)
	}
	if !empty {
		return fmt.Errorf("there are staged changes in the gitops repo. Commit them first then try again")
	}

	env := c.String("env")
	app := c.String("app")
	file := c.String("file")
	message := c.String("message")

	err = os.MkdirAll(filepath.Join(gitopsRepoPath, env, app), commands.Dir_RWX_RX_R)
	if err != nil {
		return fmt.Errorf("cannot create dir %s", err)
	}

	files, err := commands.InputFiles(file)
	if err != nil {
		return fmt.Errorf("cannot read input files %s", err)
	}
	files = splitHelmOutput(files)

	for path, content := range files {
		if !strings.HasSuffix(content, "\n") {
			content = content + "\n"
		}
		err = ioutil.WriteFile(filepath.Join(gitopsRepoPath, env, app, filepath.Base(path)), []byte(content), commands.File_RW_RW_R)
		if err != nil {
			return fmt.Errorf("cannot write file %s", err)
		}
	}

	err = stageFolder(repo, filepath.Join(env, app))
	if err != nil {
		return err
	}

	empty, err = nothingToCommit(repo)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gitMessage := fmt.Sprintf("[Gimlet CLI write] %s/%s %s", env, app, message)
	return commit(repo, gitMessage)
}

func splitHelmOutput(input map[string]string) map[string]string {
	if len(input) != 1 {
		return input
	}

	const separator = "---\n# Source: "

	files := map[string]string{}

	for _, content := range input {
		if !strings.Contains(content, separator) {
			return input
		}

		parts := strings.Split(content, separator)
		for _, p := range parts {
			p := strings.TrimSpace(p)
			if p == "" {
				continue
			}

			filePath := strings.Split(p, "\n")[0]
			fileName := filepath.Base(filePath)
			files[fileName] = strings.Join(strings.Split(p, "\n")[1:], "\n")
		}
	}

	return files
}
