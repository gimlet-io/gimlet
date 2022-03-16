package stack

import (
	"bytes"
	"fmt"
	"io/ioutil"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/stack"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var GenerateCmd = cli.Command{
	Name:      "generate",
	Usage:     "Generates Kubernetes resources from stack.yaml",
	UsageText: `stack generate`,
	Action:    generateFunc,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
		},
	},
}

func generateFunc(c *cli.Context) error {
	stackConfigPath := c.String("config")
	if stackConfigPath == "" {
		stackConfigPath = "stack.yaml"
	}
	stackConfig, err := stack.ReadStackConfig(stackConfigPath)
	if err != nil {
		return err
	}

	err = lockVersionIfNotLocked(stackConfig, stackConfigPath)
	if err != nil {
		return fmt.Errorf("couldn't lock stack version: %s", err.Error())
	}
	checkForUpdates(stackConfig)

	err = stack.GenerateAndWriteFiles(stackConfig, stackConfigPath)
	if err != nil {
		return fmt.Errorf("could not generate and write files: %s", err.Error())
	}
	fmt.Printf("\n%v  Generated\n\n", emoji.CheckMark)

	stackDefinitionYaml, err := stack.StackDefinitionFromRepo(stackConfig.Stack.Repository)
	if err != nil {
		return fmt.Errorf("cannot get stack definition: %s", err.Error())
	}
	var stackDefinition StackDefinition
	err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
	if err != nil {
		return fmt.Errorf("cannot parse stack definition: %s", err.Error())
	}

	if stackDefinition.ChangLog != "" {
		message := markdown.Render(stackDefinition.Message, 80, 6)
		fmt.Printf("%s\n", message)
	}

	return nil
}

func checkForUpdates(stackConfig dx.StackConfig) {
	currentTagString := stack.CurrentVersion(stackConfig.Stack.Repository)
	if currentTagString != "" {
		versionsSince, err := stack.VersionsSince(stackConfig.Stack.Repository, currentTagString)
		if err != nil {
			fmt.Printf("\n%v  Cannot check for updates \n\n", emoji.Warning)
		}

		if len(versionsSince) > 0 {
			fmt.Printf("\n%v  Stack update available. Run `stack update --check` for details. \n\n", emoji.Warning)
		}
	}
}

func lockVersionIfNotLocked(stackConfig dx.StackConfig, stackConfigPath string) error {
	locked, err := stack.IsVersionLocked(stackConfig)
	if err != nil {
		return fmt.Errorf("cannot check version: %s", err.Error())
	}
	if !locked {
		latestTag, _ := stack.LatestVersion(stackConfig.Stack.Repository)
		if latestTag != "" {
			stackConfig.Stack.Repository = stackConfig.Stack.Repository + "?tag=" + latestTag

			updatedStackConfigBuffer := bytes.NewBufferString("")
			e := yaml.NewEncoder(updatedStackConfigBuffer)
			e.SetIndent(2)
			e.Encode(stackConfig)

			updatedStackConfigString := "---\n" + updatedStackConfigBuffer.String()
			err = ioutil.WriteFile(stackConfigPath, []byte(updatedStackConfigString), 0666)
			if err != nil {
				return fmt.Errorf("cannot write stack file %s", err)
			}

			fmt.Printf("%v  Stack version is locked to %s \n\n", emoji.Warning, latestTag)
		}
	}

	return nil
}
