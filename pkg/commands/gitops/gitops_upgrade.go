package gitops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
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
			Name:  "no-kustomization",
			Usage: "if you don't want to upgrade your Flux repo and folder config",
		},
		&cli.BoolFlag{
			Name:  "no-deploykey",
			Usage: "if you don't want re-generate your deploy key",
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

	repo, _ := git.PlainOpen(gitopsRepoPath)
	branch, _ := branchName(repo, gitopsRepoPath)

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are changes in the gitops repo. Commit them first then try again")
	}

	noController := c.Bool("no-controller")
	noKustomization := c.Bool("no-kustomization")
	noDeployKey := c.Bool("no-deploykey")
	singleEnv := c.Bool("single-env")
	env := c.String("env")
	_, _, _, err = gitops.GenerateManifests(
		!noController,
		true,
		env,
		singleEnv,
		gitopsRepoPath,
		!noKustomization,
		!noDeployKey,
		c.String("gitops-repo-url"),
		branch,
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v GitOps configuration upgraded at %s\n\n\n", emoji.CheckMark, filepath.Join(gitopsRepoPath, env, "flux"))

	fmt.Fprintf(os.Stderr, "%v 1) Check the git diff\n", emoji.BackhandIndexPointingRight)
	fmt.Fprintf(os.Stderr, "%v 2) Commit and push to git origin\n", emoji.BackhandIndexPointingRight)

	fmt.Fprintf(os.Stderr, "\nFlux will find the changes and apply them. Essentially upgrading itself\n")

	fmt.Fprintf(os.Stderr, "\nYay Gitops%v\n\n", emoji.RaisingHands)

	return nil
}
