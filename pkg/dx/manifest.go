package dx

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"sigs.k8s.io/yaml"
)

type Manifest struct {
	App                   string                 `yaml:"app" json:"app"`
	Env                   string                 `yaml:"env" json:"env"`
	Preview               *bool                  `yaml:"preview,omitempty" json:"preview,omitempty"`
	Namespace             string                 `yaml:"namespace" json:"namespace"`
	Deploy                *Deploy                `yaml:"deploy,omitempty" json:"deploy,omitempty"`
	Cleanup               *Cleanup               `yaml:"cleanup,omitempty" json:"cleanup,omitempty"`
	Chart                 Chart                  `yaml:"chart" json:"chart"`
	Values                map[string]interface{} `yaml:"values,omitempty" json:"values,omitempty"`
	StrategicMergePatches string                 `yaml:"strategicMergePatches,omitempty" json:"strategicMergePatches,omitempty"`
	Json6902Patches       []Json6902Patch        `yaml:"json6902Patches,omitempty" json:"json6902Patches,omitempty"`
	Manifests             string                 `yaml:"manifests,omitempty" json:"manifests,omitempty"`
	Dependencies          []Dependency           `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
}

type Json6902Patch struct {
	Patch  string `yaml:"patch" json:"patch"`
	Target Target `yaml:"target" json:"target"`
}

type Target struct {
	Group   string `yaml:"group" json:"group"`
	Version string `yaml:"version" json:"version"`
	Kind    string `yaml:"kind" json:"kind"`
	Name    string `yaml:"name" json:"name"`
}

type Chart struct {
	Repository string `yaml:"repository,omitempty" json:"repository,omitempty"`
	Name       string `yaml:"name" json:"name,omitempty"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}

type Deploy struct {
	Tag                   string    `yaml:"tag,omitempty" json:"tag,omitempty"`
	Branch                string    `yaml:"branch,omitempty" json:"branch,omitempty"`
	Event                 *GitEvent `yaml:"event,omitempty" json:"event,omitempty"`
	CommitMessagePatterns []string  `yaml:"commitMessagePatterns,omitempty" json:"commitMessagePatterns,omitempty"`
}

type Cleanup struct {
	AppToCleanup string       `yaml:"app" json:"app"`
	Event        CleanupEvent `yaml:"event" json:"event"`
	Branch       string       `yaml:"branch,omitempty" json:"branch,omitempty"`
}

type Dependency struct {
	Name   string                 `yaml:"name" json:"name"`
	Chart  Chart                  `yaml:"chart" json:"chart"`
	Values map[string]interface{} `yaml:"values,omitempty" json:"values,omitempty"`
}

func (m *Manifest) PrepPreview(ingressHost string) {
	if m.Preview == nil || !*m.Preview {
		return
	}

	m.App = strings.TrimSuffix(m.App, "-preview")
	m.App = fmt.Sprintf("%s-{{ .BRANCH | sanitizeDNSName }}", m.App)

	if m.Values == nil {
		m.Values = map[string]interface{}{}
	}

	m.Values["gitBranch"] = "{{ .BRANCH }}"

	if val, ok := m.Values["ingress"]; ok {
		ingress := val.(map[string]interface{})
		if v, ok := ingress["host"]; ok {
			host := v.(string)

			if ingressHost != "" {
				ingress["host"] = m.App + ingressHost
			} else {
				dotIndex := strings.Index(host, ".")
				if dotIndex >= 0 && dotIndex < len(host) {
					host = host[dotIndex:]
					ingress["host"] = m.App + host
				}
			}
		}
	}

	m.Deploy = &Deploy{Event: PushPtr(), Branch: "!{main,master}"}

	m.Cleanup = &Cleanup{
		AppToCleanup: m.App,
		Event:        BranchDeleted,
		Branch:       "*",
	}
}

func (m *Manifest) ResolveVars(vars map[string]string) error {
	functions := make(map[string]interface{})
	for k, v := range sprig.GenericFuncMap() {
		functions[k] = v
	}
	functions["sanitizeDNSName"] = sanitizeDNSName

	// resolving vars in vars first
	varsString, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("cannot marshal vars %s", err.Error())
	}
	tpl, err := template.New("").
		Option("missingkey=error").
		Funcs(functions).
		Parse(string(varsString))
	if err != nil {
		return err
	}

	var templated bytes.Buffer
	err = tpl.Execute(&templated, vars)
	if err != nil {
		return err
	}

	var resolvedVars map[string]string
	err = yaml.Unmarshal(templated.Bytes(), &resolvedVars)
	if err != nil {
		return err
	}

	// then resolving the manifest
	cleanupBkp := m.Cleanup
	m.Cleanup = nil // cleanup only supports the BRANCH variable, not resolving it here
	manifestString, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("cannot marshal manifest %s", err.Error())
	}

	tpl, err = template.New("").
		Option("missingkey=error").
		Funcs(functions).
		Parse(string(manifestString))
	if err != nil {
		return err
	}

	err = tpl.Execute(&templated, resolvedVars)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(templated.Bytes(), m)
	m.Cleanup = cleanupBkp // restoring Cleanup after vars are resolved
	return err
}

func (m *Manifest) Render() (string, error) {
	var templatedManifests string
	var err error
	if m.Chart.Name != "" {
		templatedManifests, err = templateChart(
			m.App, m.Namespace,
			m.Chart,
			m.Values,
		)
		if err != nil {
			return templatedManifests, fmt.Errorf("cannot template Helm chart %s", err)
		}

	}

	templatedManifests += m.Manifests
	if templatedManifests == "" {
		return templatedManifests, fmt.Errorf("no chart or raw yaml has been found")
	}

	for _, dependency := range m.Dependencies {
		renderredDep, err := templateChart(
			dependency.Name,
			m.Namespace,
			dependency.Chart,
			dependency.Values,
		)
		if err != nil {
			return templatedManifests, fmt.Errorf("cannot render dependency %s", err)
		}
		templatedManifests += renderredDep
	}

	// Check for patches
	if m.StrategicMergePatches != "" || len(m.Json6902Patches) > 0 {
		templatedManifests, err = ApplyPatches(
			m.StrategicMergePatches,
			m.Json6902Patches,
			templatedManifests,
		)
		if err != nil {
			return "", fmt.Errorf("cannot apply Kustomize patches to chart %s", err)
		}
	}

	return templatedManifests, nil
}

func (c *Cleanup) ResolveVars(vars map[string]string) error {
	cleanupPolicyString, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot marshal cleanup policy %s", err.Error())
	}

	functions := make(map[string]interface{})
	for k, v := range sprig.GenericFuncMap() {
		functions[k] = v
	}
	functions["sanitizeDNSName"] = sanitizeDNSName
	tpl, err := template.New("").
		Funcs(functions).
		Parse(string(cleanupPolicyString))
	if err != nil {
		return err
	}

	var templated bytes.Buffer
	err = tpl.Execute(&templated, vars)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(templated.Bytes(), c)
}

// adheres to the Kubernetes resource name spec:
// a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-',
// and must start and end with an alphanumeric character
// (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')
func sanitizeDNSName(str string) string {
	str = strings.ToLower(str)
	r := regexp.MustCompile("[^0-9a-z]+")
	str = r.ReplaceAllString(str, "-")
	if len(str) > 53 {
		str = str[0:53]
	}
	str = strings.TrimSuffix(str, "-")
	str = strings.TrimPrefix(str, "-")
	return str
}
