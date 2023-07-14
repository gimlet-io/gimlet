package environment

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "environment",
	Usage: "Interacts with an environment on the cluster",
	Subcommands: []*cli.Command{
		&environmentConnectCmd,
		&environmentCheckCmd,
	},
}
