package dx

import (
	"bytes"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

const bareKustomization = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- manifests.yaml
`

func ApplyPatches(strategicMergePatch string, jsonPatches []Json6902Patch, manifests string) (string, error) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile("manifests.yaml", []byte(manifests))
	if err != nil {
		return "", err
	}

	kustomization := bareKustomization

	if strategicMergePatch != "" {
		err = fSys.WriteFile("strategicMergePatches.yaml", []byte(strategicMergePatch))
		if err != nil {
			return "", err
		}

		kustomization += `
patchesStrategicMerge:
- strategicMergePatches.yaml
`
	}

	if len(jsonPatches) > 0 {
		kustomization += "patches:\n"
	}
	for _, jsonPatch := range jsonPatches {
		fileName := uuid.NewString()
		err = fSys.WriteFile(fileName, []byte(jsonPatch.Patch))
		if err != nil {
			return "", err
		}

		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode([]patch{{
			Path:   fileName,
			Target: jsonPatch.Target,
		}})
		if err != nil {
			return "", err
		}
		kustomization += b.String()
	}

	err = fSys.WriteFile("kustomization.yaml", []byte(kustomization))
	if err != nil {
		return "", err
	}

	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	resources, err := b.Run(fSys, ".")
	if err != nil {
		return "", err
	}

	var files []byte
	for _, res := range resources.Resources() {
		yaml, err := res.AsYAML()
		if err != nil {
			return "", err
		}
		delimiter := []byte("---\n")
		files = append(files, delimiter...)
		files = append(files, yaml...)

	}

	return string(files), err
}

type patch struct {
	Path   string `yaml:"path" json:"path"`
	Target Target `yaml:"target" json:"target"`
}
