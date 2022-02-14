package stack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_FilterEmptyFiles(t *testing.T) {
	generatedFiles, err := generate(map[string]string{"empty-file.yaml": ""}, map[string]interface{}{})
	assert.Nil(t, err)
	assert.Empty(t, generatedFiles)

	generatedFiles, err = generate(map[string]string{"whitespace-only.yaml": "\n"}, map[string]interface{}{})
	assert.Nil(t, err)
	assert.Empty(t, generatedFiles)
}

func Test_BasicTemplating(t *testing.T) {
	stackTemplate := map[string]string{
		"template.yaml": `
{{- if .Enabled }}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
{{- end }}
`,
	}

	generatedFiles, err := generate(stackTemplate, map[string]interface{}{"Enabled": false})
	assert.Nil(t, err)
	assert.Empty(t, generatedFiles)

	generatedFiles, err = generate(stackTemplate, map[string]interface{}{"Enabled": true})
	assert.Nil(t, err)
	assert.NotEmpty(t, generatedFiles)
}

// Test_RawStrings tests when the template has {{variable}} in it that should be resolved by golang templates,
func Test_RawStrings(t *testing.T) {
	stackTemplate := map[string]string{
		"template.yaml": "rawContent: {{`\"legendFormat\": \"{{kubernetes_node}}\"`}}",
	}

	_, err := generate(stackTemplate, map[string]interface{}{"Enabled": false})
	assert.Nil(t, err)
}

func Test_UnspecifiedVarsComparison(t *testing.T) {
	stackTemplate := map[string]string{
		"template.yaml": `{{- if eq (default "" .Vendor) "do" }}hello{{- end }}`,
	}

	_, err := generate(stackTemplate, map[string]interface{}{})
	assert.Nil(t, err)
}

func Test_cloneStackFromRepo(t *testing.T) {
	//files, err := cloneStackFromRepo("git@github.com:gimlet-io/gimlet-stack-reference.git?sha=538af1fdb42fea6da80fad4c2e406ab836351f35")
	files, err := cloneStackFromRepo("https://github.com/gimlet-io/gimlet-stack-reference.git?sha=63630b03c805ef6c4c6ba02afdfe508f250d9719")
	assert.Nil(t, err)
	assert.Equal(t, 27, len(files))
}

func Test_GenerateFromStackYaml(t *testing.T) {
	stackConfigYaml := `
stack:
  repository: "https://github.com/gimlet-io/gimlet-stack-reference.git?sha=63630b03c805ef6c4c6ba02afdfe508f250d9719"
config:
  nginx:
    enabled: true
`

	var stackConfig StackConfig
	err := yaml.Unmarshal([]byte(stackConfigYaml), &stackConfig)
	assert.Nil(t, err)

	files, err := GenerateFromStackYaml(stackConfig)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	assert.Nil(t, err)
	assert.Equal(t, 5, len(files))
}

func Test_IsVersionLocked(t *testing.T) {
	stackConfigYaml := `
stack:
  repository: "https://github.com/gimlet-io/gimlet-stack-reference.git?sha=63630b03c805ef6c4c6ba02afdfe508f250d9719"
`

	var stackConfig StackConfig
	err := yaml.Unmarshal([]byte(stackConfigYaml), &stackConfig)
	assert.Nil(t, err)

	locked, err := IsVersionLocked(stackConfig)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	assert.Nil(t, err)
	assert.True(t, locked)

	stackConfigYaml = `
stack:
  repository: "https://github.com/gimlet-io/gimlet-stack-reference.git"
`

	err = yaml.Unmarshal([]byte(stackConfigYaml), &stackConfig)
	assert.Nil(t, err)

	locked, err = IsVersionLocked(stackConfig)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	assert.Nil(t, err)
	assert.False(t, locked)
}
