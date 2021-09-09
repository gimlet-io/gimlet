package release

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "release",
	Usage: "Manages Gimlet releases",
	Subcommands: []*cli.Command{
		&releaseListCmd,
		&releaseMakeCmd,
		&releaseRollbackCmd,
		&releaseTrackCmd,
		&releaseStatusCmd,
		&releaseDeleteCmd,
	},
}
