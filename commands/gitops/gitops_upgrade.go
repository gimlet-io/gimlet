package gitops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimletd/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
)

var gitopsUpgradeCmd = cli.Command{
	Name:  "upgrade",
	Usage: "Upgrades the gitops controller for an environment",
	UsageText: `gimlet gitops upgrade \
     --env staging`,
	Action: Upgrade,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "env",
			Usage: "environment to bootstrap",
		},
		&cli.BoolFlag{
			Name:  "single-env",
			Usage: "if the repo holds manifests from a single environment",
		},
		&cli.StringFlag{
			Name:     "gitops-repo-url",
			Usage:    "URL of the gitops repo (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo, default: current dir",
		},
		&cli.BoolFlag{
			Name:  "no-controller",
			Usage: "to not bootstrap the FluxV2 gitops controller, only the GitRepository and Kustomization to add a new source",
		},
		&cli.BoolFlag{
			Name:  "upgrade-kustomization",
			Usage: "if you want to upgrade not just Flux, but the gitops repo and folder configuration. Check diff carefully before comitting the changes",
		},
	},
}

func Upgrade(c *cli.Context) error {
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
	branch, _ := branchName(err, repo, gitopsRepoPath)
	if branch == "" {
		_, err = nativeGit.Commit(repo, "Initial commit")
		if err != nil {
			return err
		}
		branch, _ = branchName(err, repo, gitopsRepoPath)
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are changes in the gitops repo. Commit them first then try again")
	}

	noController := c.Bool("no-controller")
	singleEnv := c.Bool("single-env")
	env := c.String("env")
	upgradeKustomization := c.Bool("upgrade-kustomization")
	_, _, _, err = generateManifests(
		noController,
		env,
		singleEnv,
		gitopsRepoPath,
		upgradeKustomization,
		false,
		c.String("gitops-repo-url"),
		branch,
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v GitOps configuration upgraded at %s\n\n\n", emoji.CheckMark, filepath.Join(gitopsRepoPath, env, "flux"))

	fmt.Fprintf(os.Stderr, "%v 1) Check the git diff\n", emoji.BackhandIndexPointingRight)
	fmt.Fprintf(os.Stderr, "%v 2) Commit and push to git origin\n", emoji.BackhandIndexPointingRight)

	fmt.Fprintf(os.Stderr, "\n\t Happy Gitopsing%v\n\n", emoji.ConfettiBall)

	return nil
}
