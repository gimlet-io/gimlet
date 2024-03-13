package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var artifactAddCmd = cli.Command{
	Name:  "add",
	Usage: "Adds items to a release artifact",
	UsageText: `gimlet artifact add \
     --field name=CI \
     --field url=https://jenkins.example.com/job/dev/84/display/redirect \
     -f artifact.json`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "artifact file to update",
		},
		&cli.StringFlag{
			Name:  "varsFile",
			Usage: "attach variables from file to the artifact",
		},
		&cli.StringSliceFlag{
			Name:  "field",
			Usage: "data fields to attach to the artifact item in a key=value format",
		},
		&cli.StringSliceFlag{
			Name:  "envFile",
			Usage: "a Gimlet environment file to attach to the artifact",
		},
		&cli.StringSliceFlag{
			Name:  "var",
			Usage: "variables to make available in the Gimlet environment file",
		},
	},
	Action: add,
}

func add(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("cannot read file %s", err)
	}
	var a dx.Artifact
	err = json.Unmarshal(content, &a)
	if err != nil {
		return fmt.Errorf("cannot parse artifact file %s", err)
	}

	fields := c.StringSlice("field")
	item := map[string]interface{}{}
	for _, field := range fields {
		keyValue := strings.Split(field, "=")
		item[keyValue[0]] = strings.Join(keyValue[1:], "=")

		if a.Vars == nil {
			a.Vars = map[string]string{}
		}
		a.Vars[keyValue[0]] = strings.Join(keyValue[1:], "=")

	}
	if len(item) != 0 {
		a.Items = append(a.Items, item)
	}

	envFiles := c.StringSlice("envFile")
	envs := []*dx.Manifest{}
	for _, envFile := range envFiles {
		envString, err := ioutil.ReadFile(envFile)
		if err != nil {
			return fmt.Errorf("cannot read file %s", err)
		}
		var m dx.Manifest
		err = yaml.Unmarshal(envString, &m)
		if err != nil {
			return fmt.Errorf("cannot parse environment file %s", err)
		}
		envs = append(envs, &m)
	}
	a.Environments = append(a.Environments, envs...)

	vars := c.StringSlice("var")
	context := map[string]string{}
	for _, v := range vars {
		keyValue := strings.Split(v, "=")
		context[keyValue[0]] = strings.Join(keyValue[1:], "=")
	}
	for k, v := range context {
		if a.Context == nil {
			a.Context = map[string]string{}
		}
		a.Context[k] = v
		if a.Vars == nil {
			a.Vars = map[string]string{}
		}
		a.Vars[k] = v
	}

	if c.String("varsFile") != "" {
		varsMap, err := godotenv.Read(c.String("varsFile"))
		if err != nil {
			return fmt.Errorf("cannot read variables %s", err)
		}
		for k, v := range varsMap {
			if a.Vars == nil {
				a.Vars = map[string]string{}
			}
			a.Vars[k] = v
		}
	}

	jsonString := bytes.NewBufferString("")
	e := json.NewEncoder(jsonString)
	e.SetIndent("", "  ")
	e.Encode(a)

	err = ioutil.WriteFile(c.String("file"), jsonString.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("cannot write artifact json %s", err)
	}

	return nil
}
