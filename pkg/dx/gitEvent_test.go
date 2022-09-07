package dx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func Test_GitEventFromString(t *testing.T) {
	event, err := GitEventFromString("push")
	assert.Nil(t, err)
	assert.True(t, *event == 0, "should be push event")
	event, err = GitEventFromString("tag")
	assert.Nil(t, err)
	assert.True(t, *event == 1, "should be tag event")
	event, err = GitEventFromString("pr")
	assert.Nil(t, err)
	assert.True(t, *event == 2, "should be pr event")
}

func Test_InvalidGitEventFromString(t *testing.T) {
	event, err := GitEventFromString("invalidEventString")
	assert.Equal(t, err.Error(), "wrong input")
	assert.True(t, event == nil, "should be nil")
}
