package manifest

import (
	"github.com/urfave/cli/v2"
)

var manifestTemplateCmd = cli.Command{
	Name:      "template",
	Usage:     "Templates a Gimlet manifest",
	UsageText: `gimlet manifest template \
    -f .gimlet/staging-myapp.yaml \
    -o app-manifests.yaml \
    --vars ci.env`,
	Action:    template,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Gimlet manifest file to template, or \"-\" for stdin",
		},
		&cli.StringFlag{
			Name:    "vars",
			Aliases: []string{"v"},
			Usage:   "variables file for the templating",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output file",
		},
	},
}

func template(c *cli.Context) error {
	return nil
}
