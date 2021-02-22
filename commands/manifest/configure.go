package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimlet-cli/commands/chart"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

var manifestConfigureCmd = cli.Command{
	Name:      "configure",
	Usage:     "Configures Helm chart values in a Gimlet manifest",
	UsageText: `gimlet manifest configure - f .gimlet/staging.yaml`,
	Action:    configure,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "configuring an existing manifest file",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output values file",
		},
	},
}

var values map[string]interface{}

func configure(c *cli.Context) error {
	manifestPath := c.String("file")
	manifestString, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot read manifest file")
	}

	var m dx.Manifest
	err = yaml.Unmarshal(manifestString, &m)
	if err != nil {
		return fmt.Errorf("cannot unmarshal manifest")
	}

	var tmpChartName string
	if strings.HasPrefix(m.Chart.Name, "git@") {
		tmpChartName, err = dx.CloneChartFromRepo(m, "")
		if err != nil {
			return fmt.Errorf("cannot fetch chart from git %s", err.Error())
		}
		defer os.RemoveAll(tmpChartName)
	} else {
		tmpChartName = m.Chart.Name
	}

	existingValuesJson, err := json.Marshal(m.Values)
	if err != nil {
		return fmt.Errorf("cannot marshal values %s", err.Error())
	}

	yamlBytes, err := chart.ConfigureChart(tmpChartName, m.Chart.Repository, m.Chart.Version, existingValuesJson)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlBytes, &m.Values)
	if err != nil {
		return fmt.Errorf("cannot unmarshal configured values %s", err.Error())
	}

	yamlBuff := bytes.NewBuffer([]byte(""))
	e := yaml.NewEncoder(yamlBuff)
	e.SetIndent(2)
	e.Encode(m)

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, yamlBuff.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("cannot write values file %s", err)
		}
	} else {
		fmt.Println("---")
		fmt.Println(yamlBuff.String())
	}

	return nil
}
