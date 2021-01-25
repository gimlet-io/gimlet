package artifact

import (
	"github.com/urfave/cli/v2"
)

var artifactCreateCmd = cli.Command{
	Name:      "create",
	Usage:     "Creates a release artifact",
	UsageText: `gimlet artifact create \
     > artifact.json`,
	Action:    create,
}

func create(c *cli.Context) error {
	return nil
}
