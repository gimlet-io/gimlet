package gitops

import (
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Name:  "gitops",
	Usage: "Manages the gitops repo",
	Subcommands: []*cli.Command{
		&gitopsBootstrapCmd,
		&gitopsUpgradeCmd,
	},
}
