package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/commands"
)

const validEnv = `
app: fosdem-2021
env: staging
namespace: default
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 1
  image:
    repository: ghcr.io/gimlet-io/fosdem-2021
    tag: "{{ .GITHUB_SHA }}"
`

const invalidEnv = `
app: fosdem-2021
env: staging
namespace: default
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 1
  image:
    repository: ghcr.io/gimlet-io/fosdem-2021
    tag: {{ .GITHUB_SHA }}
`

const invalidReplicaType = `
app: fosdem-2021
env: staging
namespace: default
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 'string'
`

func TestLint(t *testing.T) {
	envFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(envFile.Name())
	ioutil.WriteFile(envFile.Name(), []byte(validEnv), 0644)

	args := strings.Split("gimlet manifest lint", " ")
	args = append(args, "-f", envFile.Name())

	t.Run("Should parse a gimlet manifest", func(t *testing.T) {
		t.Timeout(20 * time.Second)
		err = commands.Run(&Command, args)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("Should fail on parse error", func(t *testing.T) {
		ioutil.WriteFile(envFile.Name(), []byte(invalidEnv), 0644)
		err = commands.Run(&Command, args)
		if err == nil {
			t.Error("Expected an error, but got nil")
		}
	})

	t.Run("Should fail schema error", func(t *testing.T) {
		t.Timeout(60 * time.Second)
		ioutil.WriteFile(envFile.Name(), []byte(invalidReplicaType), 0644)
		err = commands.Run(&Command, args)
		if err == nil {
			t.Error("Expected an error, but got nil")
		}
	})
}
