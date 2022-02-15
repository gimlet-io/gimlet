package dx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_vars(t *testing.T) {
	var a Artifact
	json.Unmarshal([]byte(`
{
  "version": {},
  "environments": [],
  "context": {
	"CI_VAR": "civalue"
  },
  "items": [
    {
      "name": "image",
      "url": "nginx"
    }
  ]
}
`), &a)

	vars := a.Vars()
	assert.Equal(t, 3, len(vars))
	assert.Equal(t, 1, len(a.Context))
}

func Test_cueEnvironmentsToManifests(t *testing.T) {

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

	artifact := &Artifact{
		CueEnvironments: []string{cueTemplate},
	}

	manifests, err := artifact.CueEnvironmentsToManifests()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(manifests))
}
