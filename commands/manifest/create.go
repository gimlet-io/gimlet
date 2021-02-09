package manifest

import (
	"bytes"
	"fmt"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"strings"

	"io/ioutil"
)

var manifestCreateCmd = cli.Command{
	Name:      "create",
	Usage:     "Creates a Gimlet manifest",
	UsageText: `gimlet manifest create \
     -f values.yaml \
     --chart onechart/onechart
     --env staging \
     --app myapp \
     --namespace my-team \
     > .gimlet/staging.yaml`,
	Action:    create,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment your application is deployed to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "app",
			Usage:    "name of the application that you configure (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "namespace",
			Aliases:  []string{"n"},
			Usage:    "the Kubernetes namespace to deploy to (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "chart",
			Aliases:  []string{"c"},
			Usage:    "Helm chart to deploy (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Helm chart values file location to include in the manifest, or \"-\" for stdin",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output manifest file",
		},
	},
}

func create(c *cli.Context) error {
	valuesPath := c.String("file")
	values := map[string]interface{}{}
	if valuesPath != "" {
		yamlString, err := ioutil.ReadFile(valuesPath)
		if err != nil {
			return fmt.Errorf("cannot read values file")
		}

		err = yaml.Unmarshal(yamlString, &values)
		if err != nil {
			return fmt.Errorf("cannot parse values")
		}
	}

	chartString := c.String("chart")
	chartLoader := action.NewShow(action.ShowChart)
	var settings = helmCLI.New()
	chartPath, err := chartLoader.ChartPathOptions.LocateChart(chartString, settings)
	if err != nil {
		return fmt.Errorf("could not load %s Helm chart", err.Error())
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("could not load %s Helm chart", err.Error())
	}

	chartParts := strings.Split(chartString, "/")
	if len(chartParts) != 2 {
		return fmt.Errorf("helm chart must be in the <repo>/<chart> format, try `helm repo ls` to find your chart")
	}

	repoName := chartParts[0]

	var helmRepo *repo.Entry
	f, err := repo.LoadFile(helmpath.ConfigPath("repositories.yaml"))
	if err != nil {
		return fmt.Errorf("cannot load Helm repositories")
	}
	for _, r := range f.Repositories {
		if r.Name == repoName {
			helmRepo = r
			break
		}
	}

	if helmRepo == nil {
		return fmt.Errorf("cannot find Helm repository %s", repoName)
	}

	generatedManifest := dx.Manifest{
		App:       c.String("app"),
		Env:       c.String("env"),
		Namespace: c.String("namespace"),
		Chart: dx.Chart{
			Repository: helmRepo.URL,
			Name:       chart.Name(),
			Version:    chart.Metadata.Version,
		},
		Values: values,
	}

	yamlString := bytes.NewBufferString("")
	e := yaml.NewEncoder(yamlString)
	e.SetIndent(2)
	e.Encode(generatedManifest)

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, yamlString.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("cannot write values file %s", err)
		}
	} else {
		fmt.Println("---")
		fmt.Println(yamlString.String())
	}

	return nil
}
