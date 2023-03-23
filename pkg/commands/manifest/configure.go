package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/chart"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

var manifestConfigureCmd = cli.Command{
	Name:      "configure",
	Usage:     "Configures Helm chart values in a Gimlet manifest",
	UsageText: `gimlet manifest configure -f .gimlet/staging.yaml`,
	Action:    configure,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "configuring an existing manifest file",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "schema",
			Aliases: []string{"s"},
			Usage:   "schema file to render, made for schema development",
		},
		&cli.StringFlag{
			Name:    "ui-schema",
			Aliases: []string{"u"},
			Usage:   "ui schema file to render, made for schema development",
		},
		&cli.StringFlag{
			Name:    "chart",
			Aliases: []string{"c"},
			Usage:   "Helm chart to deploy",
		},
	},
}

type values struct {
	App       string                 `yaml:"app" json:"app"`
	Env       string                 `yaml:"env" json:"env"`
	Namespace string                 `yaml:"namespace" json:"namespace"`
	Values    map[string]interface{} `yaml:"values" json:"values"`
}

func configure(c *cli.Context) error {
	manifestPath := c.String("file")
	manifestString, err := ioutil.ReadFile(manifestPath)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return fmt.Errorf("cannot read manifest file")
	}

	var m dx.Manifest
	if manifestString != nil {
		err = yaml.Unmarshal(manifestString, &m)
		if err != nil {
			return fmt.Errorf("cannot unmarshal manifest: %s", err)
		}
	} else {
		chartName, repoUrl, chartVersion, err := chartInfos(c.String("chart"))
		if err != nil {
			return fmt.Errorf("cannot get chart infos: %s", err)
		}
		m = dx.Manifest{
			Namespace: "default",
			Chart: dx.Chart{
				Name:       chartName,
				Repository: repoUrl,
				Version:    chartVersion,
			},
			Values: map[string]interface{}{},
		}
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

	data := map[string]interface{}{
		"app":       m.App,
		"env":       m.Env,
		"namespace": m.Namespace,
		"values":    m.Values,
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cannot marshal values %s", err.Error())
	}

	var debugSchema, debugUISchema string
	if c.String("schema") != "" {
		debugSchemaBytes, err := ioutil.ReadFile(c.String("schema"))
		if err != nil {
			return fmt.Errorf("cannot read debugSchema file")
		}
		debugSchema = string(debugSchemaBytes)
	}
	if c.String("ui-schema") != "" {
		debugUISchemaBytes, err := ioutil.ReadFile(c.String("ui-schema"))
		if err != nil {
			return fmt.Errorf("cannot read debugUISchema file")
		}
		debugUISchema = string(debugUISchemaBytes)
	}

	yamlBytes, err := chart.ConfigureChart(
		tmpChartName,
		m.Chart.Repository,
		m.Chart.Version,
		dataJson,
		debugSchema,
		debugUISchema,
	)
	if err != nil {
		return err
	}

	var values values
	err = yaml.Unmarshal(yamlBytes, &values)
	if err != nil {
		return fmt.Errorf("cannot unmarshal configured values %s", err.Error())
	}

	m.App = values.App
	m.Env = values.Env
	m.Namespace = values.Namespace
	m.Values = values.Values

	manifestString, err = yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("cannot marshal manifest")
	}

	err = ioutil.WriteFile(manifestPath, manifestString, 0666)
	if err != nil {
		return fmt.Errorf("cannot write values file %s", err)
	}
	fmt.Println("Manifest configuration succeeded")

	return nil
}

func chartInfos(chart string) (string, string, string, error) {
	var repoUrl, chartName, chartVersion string
	if chart != "" {
		if strings.HasPrefix(chart, "git@") {
			chartName = chart
		} else {
			chartString := chart
			chartLoader := action.NewShow(action.ShowChart)
			var settings = helmCLI.New()
			chartPath, err := chartLoader.ChartPathOptions.LocateChart(chartString, settings)
			if err != nil {
				return "", "", "", fmt.Errorf("could not load %s Helm chart", err.Error())
			}

			chart, err := loader.Load(chartPath)
			if err != nil {
				return "", "", "", fmt.Errorf("could not load %s Helm chart", err.Error())
			}

			chartName = chart.Name()
			chartVersion = chart.Metadata.Version

			chartParts := strings.Split(chartString, "/")
			if len(chartParts) != 2 {
				return "", "", "", fmt.Errorf("helm chart must be in the <repo>/<chart> format, try `helm repo ls` to find your chart")
			}
			repoName := chartParts[0]

			var helmRepo *repo.Entry
			f, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
			if err != nil {
				return "", "", "", fmt.Errorf("cannot load Helm repositories")
			}
			for _, r := range f.Repositories {
				if r.Name == repoName {
					helmRepo = r
					break
				}
			}

			if helmRepo == nil {
				return "", "", "", fmt.Errorf("cannot find Helm repository %s", repoName)
			}
			repoUrl = helmRepo.URL
		}
		return chartName, repoUrl, chartVersion, nil
	}

	defaultChart := config.DefaultChart()
	return defaultChart.Name, defaultChart.Repo, defaultChart.Version, nil
}
