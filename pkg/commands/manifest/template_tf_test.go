package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/commands"
	"github.com/stretchr/testify/assert"
)

func Test_tfGeneration(t *testing.T) {
	manifestString := `
app: hello
manifests: |
  ---
  hello: yo
dependencies:
- name: my-redis
  kind: terraform
  spec:
    module:
      url: https://github.com/gimlet-io/tfmodules?sha=xyz&path=azure/postgresql-flexible-server-database
    values:
      database: my-app
      user: my-app
`

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

	ioutil.WriteFile(manifestFile.Name(), []byte(manifestString), commands.File_RW_RW_R)
	args := strings.Split("gimlet manifest template", " ")
	args = append(args, "-f", manifestFile.Name())
	args = append(args, "-o", templatedFile.Name())

	err = commands.Run(&Command, args)
	if assert.NoError(t, err) {
		templated, _ := ioutil.ReadFile(templatedFile.Name())
		assert.True(t, strings.Contains(string(templated), "kind: Terraform"), "terraform resource must be generated")
		// fmt.Println(string(templated))
	}
}
