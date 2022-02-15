package dx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const cueTemplate = `
import "text/template"

_instances: [
  "first",
  "second",
]

configs: [ for instance in _instances {
  app:       template.Execute("myapp-{{ . }}", instance)
  env:       "production"
  namespace: "production"
  chart: {
    repository: "https://chart.onechart.dev"
    name:       "cron-job"
    version:    0.32
  }
  values: {
    image: {
      repository: "<account>.dkr.ecr.eu-west-1.amazonaws.com/myapp"
      tag:        "1.1.1"
    }
  }
}]
`

func Test_cueRender(t *testing.T) {
	manifests, err := RenderCueToManifests(cueTemplate)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(manifests))
}
