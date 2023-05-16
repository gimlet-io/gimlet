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
}

func Test_updatingHelmChartInGitRepo(t *testing.T) {
	raw := `app: 'gimlet-dashboard'
env: staging
namespace: 'default'
chart:
  name: https://github.com/raffle-ai/onechart.git?sha=a988d33fdff367d6f8efddfeb311b2b1c74c8ff2&path=/charts/onechart/
values: {}
`

	latestVersion := "https://github.com/raffle-ai/onechart.git?sha=abcdef&path=/charts/onechart/"

	updated := updateChartVersion(raw, latestVersion)

	var updatedManifest dx.Manifest
	yaml.Unmarshal([]byte(updated), &updatedManifest)

	assert.Equal(t, latestVersion, updatedManifest.Chart.Name)
	assert.Empty(t, latestVersion, updatedManifest.Chart.Version)
	assert.Empty(t, latestVersion, updatedManifest.Chart.Repository)
}
