package artifact

import (
	"github.com/urfave/cli/v2"
)

var artifactAddCmd = cli.Command{
	Name:      "add",
	Usage:     "Adds items to a release artifact",
	UsageText: `gimlet artifact add \
     -f artifact.json`,
	Action:    add,
}

func add(c *cli.Context) error {
	return nil
}
