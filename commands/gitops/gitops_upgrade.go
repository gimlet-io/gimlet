package gitops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/enescakir/emoji"
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
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo, default: current dir",
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

	singleEnv := c.Bool("single-env")
	env := c.String("env")
	_, _, _, err = generateManifests(
		false,
		env,
		singleEnv,
		gitopsRepoPath,
		false,
		"",
		"",
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
