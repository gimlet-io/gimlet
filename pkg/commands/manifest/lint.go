package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/urfave/cli/v2"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
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

	var m dx.Manifest
	err = yaml.Unmarshal(envString, &m)
	if err != nil {
		return err
	}

	var tmpChartName string
	if strings.HasPrefix(m.Chart.Name, "git@") {
		tmpChartName, err = dx.CloneChartFromRepo(&m, "")
		if err != nil {
			return fmt.Errorf("cannot fetch chart from git %s", err.Error())
		}
		defer os.RemoveAll(tmpChartName)
	} else {
		tmpChartName = m.Chart.Name
	}

	chartLoader := action.NewShow(action.ShowChart)
	var settings = helmCLI.New()
	chartLoader.ChartPathOptions.RepoURL = m.Chart.Repository
	chartLoader.ChartPathOptions.Version = m.Chart.Version
	chartPath, err := chartLoader.ChartPathOptions.LocateChart(tmpChartName, settings)
	if err != nil {
		return fmt.Errorf("could not load %s Helm chart", err.Error())
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("could not load %s Helm chart", err.Error())
	}

	schema := string(chart.Schema)
	if schema == "" {
		return fmt.Errorf("chart doesn't have a values.schema.json with the Helm schema defined")
	}

	valuesJson, err := json.Marshal(m.Values)
	if err != nil {
		return fmt.Errorf("cannot marshal values %s", err.Error())
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(valuesJson)

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

	return nil
}
