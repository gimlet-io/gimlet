//go:build privatechart
// +build privatechart

package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

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
	t.Run("Should template a manifest file with a private git hosted chart", func(t *testing.T) {
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

		args := strings.Split("gimlet manifest template private chart", " ")
		args = append(args, "-f", manifestFile.Name())
		args = append(args, "-o", templatedFile.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}

		templated, err := ioutil.ReadFile(templatedFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(templated), "replicas: 10") {
			t.Fatal("should set replicas")
		}
		//fmt.Println(string(templated))
	})
}
