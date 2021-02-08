package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimletd/manifest"
	"github.com/gimlet-io/gimletd/artifact"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
)

var artifactAddCmd = cli.Command{
	Name:      "add",
	Usage:     "Adds items to a release artifact",
	UsageText: `gimlet artifact add \
     --field name=CI \
     --field url=https://jenkins.example.com/job/dev/84/display/redirect \
     -f artifact.json`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "artifact file to update",
		},
		&cli.StringSliceFlag{
			Name:     "field",
			Usage:    "data fields to attach to the artifact item in a key=value format",
		},
		&cli.StringSliceFlag{
			Name:     "envFile",
			Usage:    "a Gimlet environment file to attach to the artifact",
		},
		&cli.StringSliceFlag{
			Name:     "var",
			Usage:    "variables to make available in the Gimlet environment file",
		},
	},
	Action:    add,
}

func add(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("cannot read file %s", err)
	}
	var a artifact.Artifact
	err = json.Unmarshal(content, &a)
	if err != nil {
		return fmt.Errorf("cannot parse artifact file %s", err)
	}

	fields := c.StringSlice("field")
	item := map[string]interface{}{}
	for _, field := range fields {
		keyValue := strings.Split(field, "=")
		if len(keyValue) != 2 {
			return fmt.Errorf("--field should follow a key=value format")
		}
		item[keyValue[0]] = keyValue[1]
	}
	if len(item) != 0 {
		a.Items = append(a.Items, item)
	}

	envFiles := c.StringSlice("envFile")
	envs := []*manifest.Manifest{}
	for _, envFile := range envFiles {
		envString, err := ioutil.ReadFile(envFile)
		if err != nil {
			return fmt.Errorf("cannot read file %s", err)
		}
		var m manifest.Manifest
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
		if len(keyValue) != 2 {
			return fmt.Errorf("--var should follow a key=value format")
		}
		context[keyValue[0]] = keyValue[1]
	}
	a.Context = context

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
