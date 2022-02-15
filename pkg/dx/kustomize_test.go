package dx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ApplyStrategicMergePatches(t *testing.T) {

	strategicMergePatch := `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: my-team
spec:
  template:
    spec:
      containers:
      - name: myapp
        volumeMounts:
        - name: azure-file
          mountPath: /azure-bucket
    volumes:
    - name: azure-file
      azureFile:
        secretName: my-azure-secret
        shareName: my-azure-share
        readOnly: false
`

	manifest := `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: my-team
spec:
  selector:
  matchLabels:
    app.kubernetes.io/instance: myapp
    app.kubernetes.io/name: onechart
  template:
    spec:
      containers:
      - image: myapp:abcdef
        imagePullPolicy: IfNotPresent
        name: myapp
`

	patched, err := ApplyPatches(strategicMergePatch, []Json6902Patch{}, manifest)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(patched, "azureFile"))
}

func Test_ApplyJsonPatches(t *testing.T) {

	jsonPatch := Json6902Patch{
		Target: Target{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
			Name:    "myapp",
		},
		Patch: `
---
- op: replace
  path: /spec/template/spec/containers/0/name
  value: myapp-replaced
- op: replace
  path: /spec/template/spec/containers/0/imagePullPolicy
  value: Always
`,
	}

	manifest := `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: my-team
spec:
  selector:
  matchLabels:
    app.kubernetes.io/instance: myapp
    app.kubernetes.io/name: onechart
  template:
    spec:
      containers:
      - image: myapp:abcdef
        imagePullPolicy: IfNotPresent
        name: myapp
`

	patched, err := ApplyPatches("", []Json6902Patch{jsonPatch}, manifest)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(patched, "myapp-replaced"))
	assert.True(t, strings.Contains(patched, "Always"))
}
