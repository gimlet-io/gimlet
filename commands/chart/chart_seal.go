package chart

import (
	"fmt"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/urfave/cli/v2"
)

var chartSealCmd = cli.Command{
	Name:      "seal",
	Usage:     "Seals secrets in the manifest",
	UsageText: `gimlet chart seal -f values.yaml -o values.yaml -p .sealedSecrets`,
	Action:    seal,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "manifest file,folder or \"-\" for stdin (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "path",
			Aliases:  []string{"p"},
			Usage:    "json path to seal (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output sealed file",
		},
	},
}

func seal(c *cli.Context) error {
	file := c.String("file")
	//jsonPath := c.String("path")
	//outputPath := c.String("output")

	files, err := commands.InputFiles(file)
	if err != nil {
		return err
	}
	for path, contents := range files {
		fmt.Println(path, contents)
	}

	return nil
}
