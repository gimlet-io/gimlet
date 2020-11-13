package chart

import (
	"github.com/urfave/cli/v2"
)

var chartConfigureCmd = cli.Command{
	Name:      "configure",
	Usage:     "Configures Helm chart values",
	ArgsUsage: "<repo/name>",
	Action:    configure,
}

func configure(c *cli.Context) error {
	return nil
}
