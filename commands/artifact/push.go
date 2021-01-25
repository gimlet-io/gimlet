package artifact

import (
	"github.com/urfave/cli/v2"
)

var artifactPushCmd = cli.Command{
	Name:      "push",
	Usage:     "Pushes a release artifact to GimletD",
	UsageText: `gimlet artifact push \
     -f artifact.json
     --api-token c012367f6e6f71de17ae4c6a7baac2e9`,
	Action:    push,
}

func push(c *cli.Context) error {
	return nil
}
