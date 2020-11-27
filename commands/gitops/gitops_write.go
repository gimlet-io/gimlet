package gitops

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const file_RW_RW_R = 0664
const dir_RWX_RX_R = 0754

var gitopsWriteCmd = cli.Command{
	Name:      "write",
	Usage:     "Writes app manifests to a gitops environment",
	UsageText: `gimlet gitops write -f my-app.yaml \
     --env staging \
     --app my-app \
     -m "Releasing Bugfix 345"`,
	Action:    write,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "manifest file,folder or \"-\" for stdin to write (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "env",
			Usage: "environment to write to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "app",
			Usage: "name of the application that you configure (mandatory)",
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

func write(c *cli.Context) error {
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
		return fmt.Errorf("%s is not a git repository", gitopsRepoPath)
	}

	empty, err := nothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are staged changes in the gitops repo. Commit them first then try again")
	}

	env := c.String("env")
	app := c.String("app")
	file := c.String("file")
	message := c.String("message")

	err = os.MkdirAll(filepath.Join(gitopsRepoPath, env, app), dir_RWX_RX_R)
	if err != nil {
		return err
	}

	if strings.TrimSpace(file) == "-" {
		contents, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(gitopsRepoPath, env, app, "manifest.yaml"), contents, file_RW_RW_R)
		if err != nil {
			return err
		}
	} else {
		file, err = filepath.Abs(file)
		if err != nil {
			return err
		}
		fd, err := os.Stat(file)
		if err != nil {
			return err
		}
		if fd.IsDir() {
			dir, err := ioutil.ReadDir(file)
			if err != nil {
				return err
			}
			for _, f := range dir {
				err = copy(filepath.Join(file, f.Name()), filepath.Join(gitopsRepoPath, env, app, filepath.Base(f.Name())))
				if err != nil {
					return err
				}
			}
		} else {
			err = copy(file, filepath.Join(gitopsRepoPath, env, app, filepath.Base(file)))
			if err != nil {
				return err
			}
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
	return commit(repo, gitMessage, env, app)
}

func stageFolder(repo *git.Repository, folder string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	return worktree.AddWithOptions(&git.AddOptions{
		Glob: folder + "/*",
	})
}

func copy(from string, to string) error {
	contents, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(to, contents, file_RW_RW_R)
}
