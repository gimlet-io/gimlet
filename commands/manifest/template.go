package manifest

import (
	"fmt"

	"github.com/gimlet-io/gimletd/dx"
	"github.com/gimlet-io/gimletd/dx/helm"
	"github.com/gimlet-io/gimletd/dx/kustomize"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

	manifestPath := c.String("file")
	manifestString, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot read manifest file: %s", err.Error())
	}

	var m dx.Manifest
	err = yaml.Unmarshal(manifestString, &m)
	if err != nil {
		return fmt.Errorf("cannot unmarshal manifest: %s", err.Error())
	}

	err = m.ResolveVars(vars)
	if err != nil {
		return fmt.Errorf("cannot resolve manifest vars %s", err.Error())
	}

	// Get templates manifests
	templatesManifests, err := getTemplatesManifests(m)
	if err != nil {
		return fmt.Errorf("cannot get templates manifests %s", err.Error())
	}

	// Check for patches
	if m.StrategicMergePatches != "" {
		templatesManifests, err = kustomize.ApplyPatches(m.StrategicMergePatches, templatesManifests)
		if err != nil {
			return fmt.Errorf("cannot apply Kustomize patches to chart %s", err)
		}
	}

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, []byte(templatesManifests), 0666)
		if err != nil {
			return fmt.Errorf("cannot write values file %s", err)
		}
	} else {
		fmt.Println(templatesManifests)
	}

	return nil
}

func getTemplatesManifests(m dx.Manifest) (string, error) {

	templatemanifests, err := templateChart(m)
	if err != nil {
		return templatemanifests, fmt.Errorf("cannot template Helm chart %s", err)
	}

	templatemanifests += m.Manifests
	if templatemanifests == "" {
		return templatemanifests, fmt.Errorf("no chart or raw yaml has been found")
	}
	return templatemanifests, nil

}

func templateChart(m dx.Manifest) (string, error) {
	var templatesManifests string

	if m.Chart.Name == "" {
		return "", nil
	}

	if strings.HasPrefix(m.Chart.Name, "git@") ||
		strings.Contains(m.Chart.Name, ".git") { // for https:// git urls
		tmpChartDir, err := helm.CloneChartFromRepo(m, "")
		if err != nil {
			fmt.Errorf("cannot fetch chart from git %s", err.Error())
		}
		m.Chart.Name = tmpChartDir
		defer os.RemoveAll(tmpChartDir)

	}

	templatesManifests, err := helm.HelmTemplate(m)
	if err != nil {
		fmt.Errorf("cannot template Helm chart %s", err)
	}

	return templatesManifests, err

}
