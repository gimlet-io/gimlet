//go:build privatechart
// +build privatechart

package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
)

const manifestWithPrivateGitRepo = `
app: myapp
env: staging
namespace: my-team
chart:
  name: git@github.com:gimlet-io/onechart.git?sha=8e52597ae4fb4ed7888c819b3c77331622136aba&path=/charts/onechart/
values:
  replicas: 10
`

func Test_template_private_chart(t *testing.T) {
	g := goblin.Goblin(t)

	args := strings.Split("gimlet manifest template private chart", " ")

	g.Describe("gimlet manifest template", func() {
		g.It("Should template a manifest file with a private git hosted chart", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithPrivateGitRepo), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "replicas: 10")).IsTrue("should set replicas")
			//fmt.Println(string(templated))
		})
	})
}
