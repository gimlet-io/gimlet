package worker

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_updatingHelmChart(t *testing.T) {
	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.39.0
values: {}
`

	latestVersion := "0.47.0"

	updated := updateChartVersion(raw, latestVersion)

	var updatedManifest dx.Manifest
	yaml.Unmarshal([]byte(updated), &updatedManifest)

	assert.Equal(t, latestVersion, updatedManifest.Chart.Version)
	assert.Equal(t, "onechart", updatedManifest.Chart.Name)
}

func Test_updatingHelmChartInGitRepoHTTPSScheme(t *testing.T) {
	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  name: https://github.com/my-fork/onechart.git?sha=a988d33fdff367d6f8efddfeb311b2b1c74c8ff2&path=/charts/onechart/
values: {}
`

	latestVersion := "https://github.com/my-fork/onechart.git?sha=abcdef&path=/charts/onechart/"

	updated := updateChartVersion(raw, latestVersion)

	var updatedManifest dx.Manifest
	yaml.Unmarshal([]byte(updated), &updatedManifest)

	assert.Equal(t, latestVersion, updatedManifest.Chart.Name)
	assert.Empty(t, updatedManifest.Chart.Version)
	assert.Empty(t, updatedManifest.Chart.Repository)
}

func Test_updatingHelmChartInGitRepoSSHScheme(t *testing.T) {
	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  name: git@github.com:gimlet-io/onechart.git?sha=a988d33fdff367d6f8efddfeb311b2b1c74c8ff2&path=/charts/onechart/
values: {}
`

	latestVersion := "git@github.com:gimlet-io/onechart.git?sha=abcdef&path=/charts/onechart/"

	updated := updateChartVersion(raw, latestVersion)

	var updatedManifest dx.Manifest
	yaml.Unmarshal([]byte(updated), &updatedManifest)

	assert.Equal(t, latestVersion, updatedManifest.Chart.Name)
	assert.Empty(t, updatedManifest.Chart.Version)
	assert.Empty(t, updatedManifest.Chart.Repository)
}

func Test_updatingOnlyTheHashInHelmChartGitRepo(t *testing.T) {
	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  name: git@github.com:gimlet-io/onechart.git?sha=a988d33fdff367d6f8efddfeb311b2b1c74c8ff2&path=/charts/cron-job/
values: {}
`

	latestVersion := "git@github.com:gimlet-io/onechart.git?sha=abcdef&path=/charts/onechart/"

	updated := updateChartVersion(raw, latestVersion)

	var updatedManifest dx.Manifest
	yaml.Unmarshal([]byte(updated), &updatedManifest)

	assert.Equal(t, "git@github.com:gimlet-io/onechart.git?sha=abcdef&path=/charts/cron-job/", updatedManifest.Chart.Name)
}

func Test_configsPerEnv(t *testing.T) {
	files := map[string]string{
		"file1": `app: 'gimlet-dashboard'
env: staging
`,
		"file2": `app: 'gimlet-dashboard2'
env: preview
`,
		"file3": `app: 'gimlet-dashboard'
env: staging
`,
	}

	expected := map[string]map[string]string{
		"staging": {
			"file1": `app: 'gimlet-dashboard'
env: staging
`,
			"file3": `app: 'gimlet-dashboard'
env: staging
`,
		},
		"preview": {
			"file2": `app: 'gimlet-dashboard2'
env: preview
`,
		},
	}

	configs, err := configsPerEnv(files)
	assert.Nil(t, err)
	assert.Equal(t, expected, configs)
}

func Test_getChartLatestVersion(t *testing.T) {
	charts := []dx.Chart{
		{
			Name:    "onechart",
			Version: "0.47.0",
		},
		{
			Name:    "static-site",
			Version: "0.57.0",
		},
	}

	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  repository: https://chart.onechart.dev
  name: static-site
  version: 0.39.0
values: {}
`

	latestVersion := findLatestVersion(raw, charts)
	assert.Equal(t, charts[1].Version, latestVersion)
}

func Test_getChartLatestVersionGitRepoHTTPSScheme(t *testing.T) {
	charts := []dx.Chart{
		{
			Name: "https://github.com/my-fork/onechart.git?sha=abcdef&path=/charts/onechart/",
		},
		{
			Name: "https://github.com/my-fork/onechart.git?sha=ghijk&path=/charts/static-site/",
		},
	}

	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  name: https://github.com/my-fork/onechart.git?sha=a988d33fdff367d6f8efddfeb311b2b1c74c8ff2&path=/charts/onechart/
values: {}
`

	latestVersion := findLatestVersion(raw, charts)
	assert.Equal(t, charts[0].Name, latestVersion)
}

func Test_NonExistingLatestVersion(t *testing.T) {
	charts := []dx.Chart{
		{
			Name: "https://github.com/my-fork/onechart.git?sha=abcdef&path=/charts/onechart/",
		},
		{
			Name: "https://github.com/my-fork/onechart.git?sha=ghijk&path=/charts/static-site/",
		},
	}

	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  repository: https://chart.onechart.dev
  name: static-site
  version: 0.30.0
values: {}
`

	latestVersion := findLatestVersion(raw, charts)
	assert.Equal(t, "", latestVersion)
}
