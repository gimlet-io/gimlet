package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/commands"
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

func Test_lint(t *testing.T) {
	t.Run("Should parse a gimlet manifest", func(t *testing.T) {
		envFile, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(envFile.Name())
		ioutil.WriteFile(envFile.Name(), []byte(validEnv), commands.File_RW_RW_R)

		args := strings.Split("gimlet manifest lint", " ")
		args = append(args, "-f", envFile.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Should fail on parse error", func(t *testing.T) {
		envFile, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(envFile.Name())
		ioutil.WriteFile(envFile.Name(), []byte(invalidEnv), commands.File_RW_RW_R)

		args := strings.Split("gimlet manifest lint", " ")
		args = append(args, "-f", envFile.Name())

		err = commands.Run(&Command, args)
		if err == nil {
			t.Fatal("Expected error on parse error, but got nil")
		}
	})

	t.Run("Should fail schema error", func(t *testing.T) {
		envFile, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(envFile.Name())
		ioutil.WriteFile(envFile.Name(), []byte(invalidReplicaType), commands.File_RW_RW_R)

		args := strings.Split("gimlet manifest lint", " ")
		args = append(args, "-f", envFile.Name())

		err = commands.Run(&Command, args)
		if err == nil {
			t.Fatal("Expected error on schema error, but got nil")
		}
	})
}
