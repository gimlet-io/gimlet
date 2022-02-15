package dx

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func Test_GitEventYaml(t *testing.T) {
	yamlStr := `
---
branch: main
event: push
`
	var deployTrigger Deploy
	err := yaml.Unmarshal([]byte(yamlStr), &deployTrigger)
	assert.Nil(t, err)
	assert.True(t, deployTrigger.Branch == "main", "should parse branch")
	assert.True(t, *deployTrigger.Event == Push, "should parse event")

	marshalled, err := yaml.Marshal(Deploy{Branch: "main", Event: PushPtr()})
	assert.Nil(t, err)
	assert.Equal(t,
		`branch: main
event: push
`, string(marshalled))
}
