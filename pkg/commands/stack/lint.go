package stack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/stack/template"
	"github.com/urfave/cli/v2"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

var LintCmd = cli.Command{
	Name:      "lint",
	Usage:     "Lints the stack.yaml file",
	UsageText: `stack lint`,
	Action:    lint,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "stack.yaml to lint",
		},
	},
}

func lint(c *cli.Context) error {
	stackConfigPath := c.String("config")
	if stackConfigPath == "" {
		stackConfigPath = "stack.yaml"
	}
	stackConfigYaml, err := ioutil.ReadFile(stackConfigPath)
	if err != nil {
		return fmt.Errorf("cannot read stack config file: %s", err.Error())
	}

	var stackConfig template.StackConfig
	err = yaml.Unmarshal(stackConfigYaml, &stackConfig)
	if err != nil {
		return fmt.Errorf("cannot parse stack config file: %s", err.Error())
	}

	stackDefinitionYaml, err := template.StackDefinitionFromRepo(stackConfig.Stack.Repository)
	if err != nil {
		return fmt.Errorf("cannot get stack definition: %s", err.Error())
	}
	var stackDefinition template.StackDefinition
	err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
	if err != nil {
		return fmt.Errorf("cannot parse stack definition: %s", err.Error())
	}

	for _, component := range stackDefinition.Components {
		if val, ok := stackConfig.Config[component.Variable]; ok {
			schemaLoader := gojsonschema.NewStringLoader(component.Schema)

			valJson, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("cannot validate json schema %s", err)
			}
			documentLoader := gojsonschema.NewBytesLoader(valJson)

			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			if err != nil {
				return fmt.Errorf("cannot validate json schema %s", err)
			}

			if !result.Valid() {
				errs := strings.Builder{}
				for _, desc := range result.Errors() {
					errs.WriteString(fmt.Sprintf("- %s\n", desc))
				}
				return fmt.Errorf("schema validation failed: \n%s", errs.String())
			}
		}
	}

	return nil
}
