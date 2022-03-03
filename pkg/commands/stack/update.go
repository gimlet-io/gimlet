package stack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var UpdateCmd = cli.Command{
	Name:      "update",
	Usage:     "Updates the stack version in stack.yaml",
	UsageText: `stack update`,
	Action:    update,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
		},
		&cli.BoolFlag{
			Name: "check",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output format, eg.: json",
		},
	},
}

func update(c *cli.Context) error {
	stackConfigPath := c.String("config")
	if stackConfigPath == "" {
		stackConfigPath = "stack.yaml"
	}
	stackConfigYaml, err := ioutil.ReadFile(stackConfigPath)
	if err != nil {
		return fmt.Errorf("cannot read stack config file: %s", err.Error())
	}

	var stackConfig dx.StackConfig
	err = yaml.Unmarshal(stackConfigYaml, &stackConfig)
	if err != nil {
		return fmt.Errorf("cannot parse stack config file: %s", err.Error())
	}

	check := c.Bool("check")

	currentTagString := CurrentVersion(stackConfig.Stack.Repository)
	latestTag, _ := LatestVersion(stackConfig.Stack.Repository)
	if latestTag == "" {
		fmt.Printf("%v  cannot find latest version\n", emoji.CrossMark)
	}
	versionsSince, err := VersionsSince(stackConfig.Stack.Repository, currentTagString)
	if err != nil {
		fmt.Printf("\n%v  Cannot check for updates \n\n", emoji.Warning)
	}

	jsonOutput := c.String("output") == "json"

	if len(versionsSince) == 0 {
		if jsonOutput {
			updateStr := bytes.NewBufferString("")
			e := json.NewEncoder(updateStr)
			e.SetIndent("", "  ")
			err = e.Encode(map[string]string{
				"status": "Already up to date",
			})
			if err != nil {
				return fmt.Errorf("cannot deserialize update status %s", err)
			}
			fmt.Println(updateStr)
		} else {
			fmt.Printf("\n%v  Already up to date \n\n", emoji.CheckMark)
			return nil
		}
	}

	if check {
		if jsonOutput {
			updateStr := bytes.NewBufferString("")
			e := json.NewEncoder(updateStr)
			e.SetIndent("", "  ")
			err = e.Encode(map[string]string{
				"status":         "Update available",
				"currentVersion": currentTagString,
				"latestVersion":  latestTag,
			})
			if err != nil {
				return fmt.Errorf("cannot deserialize update status %s", err)
			}
			fmt.Println(updateStr)
		} else {
			fmt.Printf("%v  New version available: \n\n", emoji.Books)
			err := printChangeLog(stackConfig, versionsSince)
			if err != nil {
				fmt.Printf("\n%v %s \n\n", emoji.Warning, err)
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("%v  Stack version is updating to %s... \n\n", emoji.HourglassNotDone, latestTag)
		stackConfig.Stack.Repository = RepoUrlWithoutVersion(stackConfig.Stack.Repository) + "?tag=" + latestTag
		err = writeStackConfig(stackConfig, stackConfigPath)
		if err != nil {
			return fmt.Errorf("cannot write stack file %s", err)
		}
		fmt.Printf("%v   Config updated. \n\n", emoji.CheckMark)
		fmt.Printf("%v   Run `stack generate` to render resources with the updated stack. \n\n", emoji.Warning)
		fmt.Printf("%v  Change log:\n\n", emoji.Books)
		err = printChangeLog(stackConfig, versionsSince)
		if err != nil {
			fmt.Printf("\n%v %s \n\n", emoji.Warning, err)
		}
		fmt.Printf("\n")
	}

	return nil
}

func writeStackConfig(stackConfig dx.StackConfig, stackConfigPath string) error {
	updatedStackConfigBuffer := bytes.NewBufferString("")
	e := yaml.NewEncoder(updatedStackConfigBuffer)
	e.SetIndent(2)
	e.Encode(stackConfig)

	updatedStackConfigString := "---\n" + updatedStackConfigBuffer.String()
	return ioutil.WriteFile(stackConfigPath, []byte(updatedStackConfigString), 0666)
}

func printChangeLog(stackConfig dx.StackConfig, versions []string) error {
	for _, version := range versions {
		fmt.Printf("   - %s \n", version)

		repoUrl := stackConfig.Stack.Repository
		repoUrl = RepoUrlWithoutVersion(repoUrl)
		repoUrl = repoUrl + "?tag=" + version

		stackDefinitionYaml, err := StackDefinitionFromRepo(repoUrl)
		if err != nil {
			return fmt.Errorf("cannot get stack definition: %s", err.Error())
		}
		var stackDefinition StackDefinition
		err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
		if err != nil {
			return fmt.Errorf("cannot parse stack definition: %s", err.Error())
		}

		if stackDefinition.ChangLog != "" {
			changeLog := markdown.Render(stackDefinition.ChangLog, 80, 6)
			fmt.Printf("%s\n", changeLog)
		}
	}

	return nil
}
