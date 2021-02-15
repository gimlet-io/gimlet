package manifest

import (
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

const validEnv = `
app: fosdem-2021
env: staging
namespace: default
chart:
  name: git@github.com:gimlet-io/onechart.git?sha=8e52597ae4fb4ed7888c819b3c77331622136aba&path=/charts/onechart/
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
	envFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(envFile.Name())
	ioutil.WriteFile(envFile.Name(), []byte(validEnv), commands.File_RW_RW_R)

	args := strings.Split("gimlet manifest lint", " ")
	args = append(args, "-f", envFile.Name())

	g := goblin.Goblin(t)
	g.Describe("gimlet manifest lint", func() {
		g.It("Should parse a gimlet manifest", func() {
			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)
		})
		g.It("Should fail on parse error", func() {
			ioutil.WriteFile(envFile.Name(), []byte(invalidEnv), commands.File_RW_RW_R)
			err = commands.Run(&Command, args)
			g.Assert(err != nil).IsTrue(err)
		})
		g.It("Should fail schema error", func() {
			g.Timeout(60*time.Second)
			ioutil.WriteFile(envFile.Name(), []byte(invalidReplicaType), commands.File_RW_RW_R)
			err = commands.Run(&Command, args)
			g.Assert(err != nil).IsTrue(err)
		})
	})
}
