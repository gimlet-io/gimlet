package manifest

import (
	"fmt"

	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"io/ioutil"
	"os"
	"strings"
)

var manifestTemplateCmd = cli.Command{
	Name:  "template",
	Usage: "Templates a Gimlet manifest",
	UsageText: `gimlet manifest template \
    -f .gimlet/staging.yaml \
    -o manifests.yaml \
    --vars ci.env`,
	Action: templateCmd,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Required: true,
			Usage:    "Gimlet manifest file to template, or \"-\" for stdin",
		},
		&cli.StringFlag{
			Name:    "vars",
			Aliases: []string{"v"},
			Usage:   "an .env file for template variables",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output file",
		},
	},
}

func templateCmd(c *cli.Context) error {
	varsPath := c.String("vars")
	vars := map[string]string{}
	if varsPath != "" {
		yamlString, err := ioutil.ReadFile(varsPath)
		if err != nil {
			return fmt.Errorf("cannot read vars file: %s", err.Error())
		}

		vars, err = godotenv.Parse(strings.NewReader(string(yamlString)))
		if err != nil {
			return fmt.Errorf("cannot parse vars: %s", err.Error())
		}
	}

	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		if _, exists := vars[pair[0]]; !exists {
			vars[pair[0]] = pair[1]
		}
	}

	var templatedManifests string

	filePath := c.String("file")
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %s", err.Error())
	}

	if strings.HasSuffix(filePath, ".cue") { // handling CUE format
		manifests, err := dx.RenderCueToManifests(string(fileContent))
		if err != nil {
			return fmt.Errorf("cannot parse cue file: %s", err.Error())
		}

		for _, m := range manifests {
			tm, err := parseResolveAndRenderManifest([]byte(m), vars)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			templatedManifests += tm
		}
	} else { // handling YAML format
		templatedManifests, err = parseResolveAndRenderManifest(fileContent, vars)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, []byte(templatedManifests), 0666)
		if err != nil {
			return fmt.Errorf("cannot write values file %s", err)
		}
	} else {
		fmt.Println(templatedManifests)
	}

	return nil
}

func parseResolveAndRenderManifest(manifestString []byte, vars map[string]string) (string, error) {
	var m dx.Manifest
	err := yaml.Unmarshal(manifestString, &m)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal manifest: %s", err.Error())
	}

	m.PrepPreview("")
	err = m.ResolveVars(vars)
	if err != nil {
		return "", fmt.Errorf("cannot resolve manifest vars %s", err.Error())
	}

	return m.Render()
}
