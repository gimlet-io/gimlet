package chart

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "chart",
	Usage: "Manages Helm charts",
	Subcommands: []*cli.Command{
		&chartConfigureCmd,
	},
}
