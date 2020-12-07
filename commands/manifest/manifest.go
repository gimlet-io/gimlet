package manifest

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "manifest",
	Usage: "Manages Gimlet manifests",
	Subcommands: []*cli.Command{
		&manifestCreateCmd,
	},
}
