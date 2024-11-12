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
