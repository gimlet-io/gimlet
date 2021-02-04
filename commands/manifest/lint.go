package manifest

import (
	"fmt"
	"github.com/gimlet-io/gimletd/manifest"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var manifestLintCmd = cli.Command{
	Name:      "lint",
	Usage:     "Lints a Gimlet manifest",
	UsageText: `gimlet manifest lint -f .gimlet/staging.yaml`,
	Action:    lint,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "Gimlet manifest file to lint",
			Required: true,
		},
	},
}

func lint(c *cli.Context) error {
	envFile := c.String("file")
	envString, err := ioutil.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("cannot read file %s", err)
	}

	var m manifest.Manifest
	err = yaml.Unmarshal(envString, &m)
	if err != nil {
		return err
	}

	return nil
}
