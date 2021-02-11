package manifest

import (
	"fmt"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	giturl "github.com/whilp/git-urls"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
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
			return fmt.Errorf("cannot read vars file")
		}

		vars, err = godotenv.Parse(strings.NewReader(string(yamlString)))
		if err != nil {
			return fmt.Errorf("cannot parse vars")
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
		return fmt.Errorf("cannot read manifest file")
	}

	var m dx.Manifest
	err = yaml.Unmarshal(manifestString, &m)
	if err != nil {
		return fmt.Errorf("cannot unmarshal manifest")
	}

	err = m.ResolveVars(vars)
	if err != nil {
		return fmt.Errorf("cannot resolve manifest vars %s", err.Error())
	}

	if strings.HasPrefix(m.Chart.Name, "git@") {
		gitAddress, err := giturl.ParseScp(m.Chart.Name)
		if err != nil {
			return fmt.Errorf("cannot parse chart's git address: %s", err)
		}
		gitUrl := strings.ReplaceAll(m.Chart.Name, gitAddress.RawQuery, "")
		gitUrl = strings.ReplaceAll(gitUrl, "?", "")

		tmpChartDir, err := ioutil.TempDir("", "gimlet-git-chart")
		if err != nil {
			return fmt.Errorf("cannot create tmp file: %s", err)
		}
		defer os.RemoveAll(tmpChartDir)

		repo, err := git.PlainClone(tmpChartDir, false, &git.CloneOptions{
			URL:   gitUrl,
		})
		if err != nil {
			return fmt.Errorf("cannot clone chart git repo: %s", err)
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("cannot get worktree: %s", err)
		}

		params, _ := url.ParseQuery(gitAddress.RawQuery)
		if v, found := params["path"]; found {
			tmpChartDir = tmpChartDir + v[0]
		}
		if v, found := params["sha"]; found {
			err = worktree.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(v[0]),
			})
			if err != nil {
				return fmt.Errorf("cannot checkout sha: %s", err)
			}
		}
		if v, found := params["tag"]; found {
			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewTagReferenceName(v[0]),
			})
			if err != nil {
				return fmt.Errorf("cannot checkout tag: %s", err)
			}
		}
		if v, found := params["branch"]; found {
			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName(v[0]),
			})
			if err != nil {
				return fmt.Errorf("cannot checkout branch: %s", err)
			}
		}

		m.Chart.Name = tmpChartDir
	}

	templatesManifests, err := dx.HelmTemplate(m)
	if err != nil {
		return fmt.Errorf("cannot template Helm chart %s", err)
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
