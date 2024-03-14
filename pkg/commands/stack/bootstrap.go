package stack

import (
	"github.com/gimlet-io/gimlet/pkg/commands/gitops"
	"github.com/urfave/cli/v2"
)

var BootstrapCmd = cli.Command{
	Name:      "bootstrap",
	Usage:     "Bootstraps the gitops controller",
	UsageText: `stack bootstrap --single-env --gitops-repo-url git@github.com:<user>/<repo>.git`,
	Action:    gitops.Bootstrap,
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
	},
}
