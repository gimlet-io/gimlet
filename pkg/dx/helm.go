package dx

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	giturl "github.com/whilp/git-urls"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
)

// SplitHelmOutput splits helm's multifile string output into file paths and their content
func SplitHelmOutput(input map[string]string) map[string]string {
	if len(input) != 1 {
		return input
	}

	const separator = "---\n# Source: "

	files := map[string]string{}

	for _, content := range input {
		if !strings.Contains(content, separator) {
			return input
		}

		parts := strings.Split(content, separator)
		for _, p := range parts {
			p := strings.TrimSpace(p)
			if p == "" {
				continue
			}

			lines := strings.Split(p, "\n")
			filePath := lines[0]
			content := strings.Join(lines[1:], "\n")
			fileName := filepath.Base(filePath)
			if existingContent, ok := files[fileName]; ok {
				files[fileName] = existingContent + "---\n" + content + "\n"
			} else {
				files[fileName] = "---\n" + content + "\n"
			}
		}
	}

	return files
}

// CloneChartFromRepo returns the chart location of the specified chart
func CloneChartFromRepo(m *Manifest, token string) (string, error) {
	gitAddress, err := giturl.Parse(m.Chart.Name)
	if err != nil {
		return "", fmt.Errorf("cannot parse chart's git address: %s", err)
	}
	gitUrl := strings.ReplaceAll(m.Chart.Name, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	tmpChartDir, err := ioutil.TempDir("", "gimlet-git-chart")
	if err != nil {
		return "", fmt.Errorf("cannot create tmp file: %s", err)
	}

	opts := &git.CloneOptions{
		URL: gitUrl,
	}
	if token != "" {
		opts.Auth = &http.BasicAuth{
			Username: "abc123", // this can be anything
			Password: token,
		}
	}
	repo, err := git.PlainClone(tmpChartDir, false, opts)
	if err != nil {
		return "", fmt.Errorf("cannot clone chart git repo: %s", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("cannot get worktree: %s", err)
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
			return "", fmt.Errorf("cannot checkout sha: %s", err)
		}
	}
	if v, found := params["tag"]; found {
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(v[0]),
		})
		if err != nil {
			return "", fmt.Errorf("cannot checkout tag: %s", err)
		}
	}
	if v, found := params["branch"]; found {
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewRemoteReferenceName("origin", v[0]),
		})
		if err != nil {
			return "", fmt.Errorf("cannot checkout branch: %s", err)
		}
	}

	return tmpChartDir, nil
}

func ChartSchema(m *Manifest, installationToken string) (string, string, error) {
	client, settings := helmClient(m)
	chartFromManifest, err := loadChartFromManifest(m, client, settings, installationToken)
	if err != nil {
		return "", "", err
	}

	var schemaUI string
	for _, file := range chartFromManifest.Files {
		if file.Name == "helm-ui.json" {
			schemaUI = string(file.Data)
			break
		}
	}

	return string(chartFromManifest.Schema), schemaUI, nil
}

func templateChart(m *Manifest) (string, error) {
	client, settings := helmClient(m)
	chartFromManifest, err := loadChartFromManifest(m, client, settings, "")
	if err != nil {
		return "", err
	}

	rel, err := client.Run(chartFromManifest, m.Values)
	if err != nil {
		return "", err
	}

	return rel.Manifest, err

}

func loadChartFromManifest(m *Manifest, client *action.Install, settings *helmCLI.EnvSettings, token string) (*chart.Chart, error) {
	if m.Chart.Name == "" {
		return nil, nil
	}

	if strings.HasPrefix(m.Chart.Name, "git@") ||
		strings.Contains(m.Chart.Name, ".git") { // for https:// git urls
		tmpChartDir, err := CloneChartFromRepo(m, token)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch chart from git %s", err.Error())
		}
		m.Chart.Name = tmpChartDir
		defer os.RemoveAll(tmpChartDir)

	}

	cp, err := client.ChartPathOptions.LocateChart(m.Chart.Name, settings)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}

func helmClient(m *Manifest) (*action.Install, *helmCLI.EnvSettings) {
	actionConfig := new(action.Configuration)
	client := action.NewInstall(actionConfig)

	client.DryRun = true
	client.ReleaseName = m.App
	client.Replace = true
	client.ClientOnly = true
	client.APIVersions = []string{}
	client.IncludeCRDs = false
	client.ChartPathOptions.RepoURL = m.Chart.Repository
	client.ChartPathOptions.Version = m.Chart.Version
	client.Namespace = m.Namespace

	var settings = helmCLI.New()
	return client, settings
}
