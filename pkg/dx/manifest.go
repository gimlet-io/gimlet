package dx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	terraformv1 "github.com/weaveworks/tf-controller/api/v1alpha2"
	giturl "github.com/whilp/git-urls"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Name       string `yaml:"name" json:"name"`
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
	Name string      `yaml:"name" json:"name"`
	Kind string      `yaml:"kind" json:"kind"`
	Spec interface{} `yaml:"spec" json:"spec"`
}

func (d *Dependency) UnmarshalJSON(data []byte) error {
	var dat map[string]interface{}

	if err := json.Unmarshal(data, &dat); err != nil {
		return nil
	}

	d.Name = dat["name"].(string)
	if _, ok := dat["kind"]; !ok {
		return fmt.Errorf("kind is mandatory for dependency")
	}
	d.Kind = dat["kind"].(string)

	switch d.Kind {
	case "terraform":
		dat = dat["spec"].(map[string]interface{})
		module := dat["module"].(map[string]interface{})
		tfSpec := TFSpec{
			Module: Module{
				Url: module["url"].(string),
			},
			Values: dat["values"].(map[string]interface{}),
		}
		if val, ok := module["secret"]; ok {
			tfSpec.Module.Secret = val.(string)
		}
		if val, ok := dat["secret"]; ok {
			tfSpec.Secret = val.(string)
		}
		d.Spec = tfSpec
	}
	return nil
}

type TFSpec struct {
	Module Module                 `yaml:"module" json:"module"`
	Values map[string]interface{} `yaml:"values" json:"values"`
	Secret string                 `yaml:"secret" json:"secret"`
}

type Module struct {
	Url    string `yaml:"url" json:"url"`
	Secret string `yaml:"secret,omitempty" json:"secret,omitempty"`
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
		templatedManifests, err = templateChart(m)
		if err != nil {
			return templatedManifests, fmt.Errorf("cannot template Helm chart %s", err)
		}

	}

	templatedManifests += m.Manifests
	if templatedManifests == "" {
		return templatedManifests, fmt.Errorf("no chart or raw yaml has been found")
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

	for _, dependency := range m.Dependencies {
		renderredDep, err := renderDependency(dependency, m)
		if err != nil {
			return templatedManifests, fmt.Errorf("cannot render dependency %s", err)
		}
		templatedManifests += renderredDep
	}

	return templatedManifests, nil
}

func renderDependency(dependency Dependency, manifest *Manifest) (string, error) {
	depString := ""
	switch dependency.Kind {
	case "terraform":
		tfSpec := dependency.Spec.(TFSpec)

		gitAddress, err := giturl.Parse(tfSpec.Module.Url)
		if err != nil {
			return "", fmt.Errorf("cannot parse dependency's git address: %s", err)
		}
		moduleUrl := strings.ReplaceAll(tfSpec.Module.Url, gitAddress.RawQuery, "")
		moduleUrl = strings.ReplaceAll(moduleUrl, "?", "")

		params, _ := url.ParseQuery(gitAddress.RawQuery)
		branch := ""
		if v, found := params["branch"]; found {
			branch = v[0]
		}
		tag := ""
		if v, found := params["tag"]; found {
			tag = v[0]
		}
		sha := ""
		if v, found := params["sha"]; found {
			sha = v[0]
		}
		path := ""
		if v, found := params["path"]; found {
			path = v[0]
		}

		gitRepoBytes, err := renderTFGitRepo(
			manifest.App+"-"+dependency.Name,
			manifest.Namespace,
			moduleUrl,
			branch,
			tag,
			sha,
			tfSpec.Module.Secret,
		)
		if err != nil {
			return "", err
		}
		depString += "---\n"
		depString += string(gitRepoBytes)

		tfKindBytes, err := renderTFKind(
			manifest.App+"-"+dependency.Name,
			manifest.Namespace,
			moduleUrl,
			branch,
			tag,
			sha,
			path,
			tfSpec.Secret,
			tfSpec.Values,
		)
		if err != nil {
			return "", err
		}
		depString += "---\n"
		depString += string(tfKindBytes)

	}
	return depString, nil
}

func renderTFGitRepo(
	name string,
	namespace string,
	url string,
	branch string,
	tag string,
	sha string,
	secretName string,
) ([]byte, error) {
	gvk := sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	gitRepository := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: url,
			Interval: metav1.Duration{
				Duration: 24 * time.Hour,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: branch,
				Tag:    tag,
				Commit: sha,
			},
		},
	}

	if secretName != "" {
		gitRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: secretName,
		}
	}

	return yaml.Marshal(gitRepository)
}

func renderTFKind(
	name string,
	namespace string,
	url string,
	branch string,
	tag string,
	sha string,
	path string,
	secretName string,
	vars map[string]interface{},
) ([]byte, error) {
	terraform := terraformv1.Terraform{
		TypeMeta: metav1.TypeMeta{
			Kind:       terraformv1.TerraformKind,
			APIVersion: terraformv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: terraformv1.TerraformSpec{
			Interval: metav1.Duration{
				Duration: 24 * time.Hour,
			},
			ApprovePlan: terraformv1.ApprovePlanAutoValue,
			Path:        path,
			SourceRef: terraformv1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
				Name: name,
			},
			WriteOutputsToSecret: &terraformv1.WriteOutputsToSecretSpec{
				Name: name + "-output",
			},
			VarsFrom: []terraformv1.VarsReference{
				{
					Kind: "Secret",
					Name: secretName,
				},
			},
			Vars: []terraformv1.Variable{},
		},
	}

	for k, v := range vars {
		terraform.Spec.Vars = append(
			terraform.Spec.Vars,
			terraformv1.Variable{
				Name:  k,
				Value: &v1.JSON{Raw: []byte("\"" + v.(string) + "\"")},
			},
		)
	}

	return yaml.Marshal(terraform)
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
