package dx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

func Test_resolveVars(t *testing.T) {
	m := &Manifest{
		App:       "my-app",
		Namespace: "my-namespace",
		Values: map[string]interface{}{
			"image": "debian",
		},
	}

	err := m.ResolveVars(map[string]string{})
	assert.Nil(t, err)
	assert.Equal(t, "my-app", m.App)

	m = &Manifest{
		App:       "my-app-{{ .POSTFIX }}",
		Namespace: "my-namespace",
		Values: map[string]interface{}{
			"image": "debian:{{ .POSTFIX }}",
		},
	}

	err = m.ResolveVars(map[string]string{
		"POSTFIX": "test",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-app-test", m.App)
	assert.Equal(t, "debian:test", m.Values["image"])

	m = &Manifest{
		App:       "my-app-{{ .BRANCH | sanitizeDNSName }}",
		Namespace: "my-namespace",
		Values: map[string]interface{}{
			"image": "debian:{{ .BRANCH | sanitizeDNSName }}",
		},
	}

	err = m.ResolveVars(map[string]string{
		"BRANCH": "feature/my-feature",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-app-feature-my-feature", m.App)
	assert.Equal(t, "debian:feature-my-feature", m.Values["image"])
}

func Test_sanitizeDNSName(t *testing.T) {
	sanitized := sanitizeDNSName("CamelCase_with_snake")
	assert.Equal(t, "camelcase-with-snake", sanitized)

	sanitized = sanitizeDNSName("dependabot/npm_and_yarn/ws-5.2.3")
	assert.Equal(t, "dependabot-npm-and-yarn-ws-5-2-3", sanitized)

	sanitized = sanitizeDNSName("-can't start with dashes, nor end-")
	assert.Equal(t, "can-t-start-with-dashes-nor-end", sanitized)

	sanitized = sanitizeDNSName("!nope")
	assert.Equal(t, "nope", sanitized)

	sanitized = sanitizeDNSName("dope")
	assert.Equal(t, "dope", sanitized)
}

func Test_resolveVars_missingVar(t *testing.T) {

	m := &Manifest{
		App:       "my-app-{{ .POSTFIX }}",
		Namespace: "my-namespace",
	}

	err := m.ResolveVars(map[string]string{})
	assert.True(t, strings.Contains(err.Error(), "map has no entry for key \"POSTFIX\""))
}

func Test_UnmarshallingToDefaultValues(t *testing.T) {
	manifestString := ""

	var m Manifest
	err := yaml.Unmarshal([]byte(manifestString), &m)
	assert.Nil(t, err)

	manifestString = `
app: hello
`

	err = yaml.Unmarshal([]byte(manifestString), &m)
	assert.Nil(t, err)
	assert.Equal(t, "hello", m.App)

	manifestString = `
app: hello
`

	err = v2.Unmarshal([]byte(manifestString), &m)
	assert.Nil(t, err)
	assert.Equal(t, "hello", m.App)
}

func Test_dependencyUnmarshal(t *testing.T) {
	manifestString := `
app: hello
dependencies:
- name: my-redis
  kind: terraform
  spec:
    module:
      url: https://github.com/gimlet-io/tfmodules?tag=v1.0.0&path=aws/elasticache
    values:
      size: 1GB
    secret: xx
`

	var m Manifest
	err := yaml.Unmarshal([]byte(manifestString), &m)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(m.Dependencies))
		dep := m.Dependencies[0]
		assert.Equal(t, "my-redis", dep.Name)
		assert.Equal(t, "terraform", dep.Kind)
		tfSpec := dep.Spec.(TFSpec)
		assert.Equal(t, "https://github.com/gimlet-io/tfmodules?tag=v1.0.0&path=aws/elasticache", tfSpec.Module.Url)
		assert.Equal(t, "xx", tfSpec.Secret)
	}
}

func Test_dependencyMarshal(t *testing.T) {
	m := Manifest{
		App: "first",
		Dependencies: []Dependency{
			{
				Name: "my-redis",
				Kind: "terraform",
				Spec: TFSpec{
					Module: Module{
						Url: "a-git-url",
					},
					Values: map[string]interface{}{
						"size": "1GB",
					},
					Secret: "xx",
				},
			},
		},
	}

	marshalledBytes, err := yaml.Marshal(m)
	if assert.NoError(t, err) {
		assert.Equal(t, `app: first
chart:
  name: ""
dependencies:
- kind: terraform
  name: my-redis
  spec:
    module:
      url: a-git-url
    secret: xx
    values:
      size: 1GB
env: ""
namespace: ""
`, string(marshalledBytes))
	}
}

func Test_planiDependencyMarshal(t *testing.T) {
	m := Manifest{
		App: "first",
		Dependencies: []Dependency{
			{
				Name: "my-redis",
				Kind: "plain",
				Spec: TFSpec{
					Module: Module{
						Url: "a-git-url",
					},
					Values: map[string]interface{}{
						"size": "1GB",
					},
					Secret: "xx",
				},
			},
		},
	}

	marshalledBytes, err := yaml.Marshal(m)
	if assert.NoError(t, err) {
		assert.Equal(t, `app: first
chart:
  name: ""
dependencies:
- kind: plain
  name: my-redis
  spec:
    module:
      url: a-git-url
    secret: xx
    values:
      size: 1GB
env: ""
namespace: ""
`, string(marshalledBytes))
	}
}

func Test_renderTFDependency(t *testing.T) {
	manifestString := `
app: hello
manifests: |
  ---
  hello: yo
dependencies:
- name: my-redis
  kind: terraform
  spec:
    module:
      url: https://github.com/gimlet-io/tfmodules?sha=xyz&path=azure/postgresql-flexible-server-database
      secret: gitDeployKey
    values:
      database: my-app
      user: my-app
    secret: db-admin-secret
`

	var m Manifest
	err := yaml.Unmarshal([]byte(manifestString), &m)
	if assert.NoError(t, err) {
		renderredDep, err := renderDependency(m.Dependencies[0], &m)
		if assert.NoError(t, err) {
			assert.True(t, strings.Contains(string(renderredDep), "url: https://github.com/gimlet-io/tfmodule"), "git repo url must be set")
			assert.True(t, strings.Contains(string(renderredDep), "commit: xyz"), "git tag must be set")
			assert.True(t, strings.Contains(string(renderredDep), "kind: Terraform"), "terraform kind must be set")
			assert.True(t, strings.Contains(string(renderredDep), "name: db-admin-secret"), "db secret must be set")
			assert.True(t, strings.Contains(string(renderredDep), "value: my-app"), "values must be set")
			// fmt.Println(string(renderredDep))
		}
	}
}

func Test_PrepPreview(t *testing.T) {
	notPreview := &Manifest{
		App:       "my-app",
		Namespace: "my-namespace",
		Values: map[string]interface{}{
			"image": "debian",
		},
	}

	notPreview.PrepPreview("")
	assert.Equal(t, "my-app", notPreview.App)

	boolTrue := true
	preview := &Manifest{
		App:       "my-app-preview",
		Namespace: "my-namespace",
		Preview:   &boolTrue,
		Values: map[string]interface{}{
			"image": "debian",
		},
	}

	preview.PrepPreview("")
	assert.Equal(t, "my-app-{{ .BRANCH | sanitizeDNSName }}", preview.App)

	previewWithIngress := &Manifest{
		App:       "my-app-preview",
		Namespace: "my-namespace",
		Preview:   &boolTrue,
		Values: map[string]interface{}{
			"image": "debian",
			"ingress": map[string]interface{}{
				"host": "my-app-preview.gimlet.app",
			},
		},
	}

	previewWithIngress.PrepPreview("")
	ingressValues := previewWithIngress.Values["ingress"].(map[string]interface{})
	assert.Equal(t, "my-app-{{ .BRANCH | sanitizeDNSName }}.gimlet.app", ingressValues["host"])
	assert.Equal(t, "{{ .BRANCH }}", previewWithIngress.Values["gitBranch"])

	assert.NotNil(t, preview.Deploy)
	push := Push
	assert.Equal(t, &push, preview.Deploy.Event)
	assert.Equal(t, "!{main,master}", preview.Deploy.Branch)

	assert.NotNil(t, preview.Cleanup)
	assert.Equal(t, "my-app-{{ .BRANCH | sanitizeDNSName }}", preview.Cleanup.AppToCleanup)
}

func Test_PrepPreview_Ingress(t *testing.T) {
	boolTrue := true

	previewWithIngress := &Manifest{
		App:       "my-app-preview",
		Namespace: "my-namespace",
		Preview:   &boolTrue,
		Values: map[string]interface{}{
			"image": "debian",
			"ingress": map[string]interface{}{
				"host": "my-app-preview.gimlet.app",
			},
		},
	}

	previewWithIngress.PrepPreview("")
	ingressValues := previewWithIngress.Values["ingress"].(map[string]interface{})
	assert.Equal(t, "my-app-{{ .BRANCH | sanitizeDNSName }}.gimlet.app", ingressValues["host"])

	previewWithIngress = &Manifest{
		App:       "my-app-preview",
		Namespace: "my-namespace",
		Preview:   &boolTrue,
		Values: map[string]interface{}{
			"image": "debian",
			"ingress": map[string]interface{}{
				"host": "my-app-preview-blabla.gimlet.app",
			},
		},
	}

	previewWithIngress.PrepPreview("-blabla.gimlet.app")
	ingressValues = previewWithIngress.Values["ingress"].(map[string]interface{})
	assert.Equal(t, "my-app-{{ .BRANCH | sanitizeDNSName }}-blabla.gimlet.app", ingressValues["host"])

}
