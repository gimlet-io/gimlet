package artifact

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "artifact",
	Usage: "Manages release artifacts",
	Subcommands: []*cli.Command{
		&artifactCreateCmd,
		&artifactAddCmd,
		&artifactPushCmd,
		&artifactListCmd,
		&artifactTrackCmd,
	},
}
