package stack

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/stretchr/testify/assert"
)

const valid = `
---
stack:
  repository: https://github.com/gimlet-io/gimlet-stack-reference.git
config:
  nginx:
    enabled: true
    host: "gimlet.io"
`

const invalidFieldType = `
---
stack:
  repository: https://github.com/gimlet-io/gimlet-stack-reference.git
config:
  nginx:
    enabled: "true"
    host: "gimlet.io"
`

func TestLint(t *testing.T) {
	stackFile, err := ioutil.TempFile("", "stack-test")
	if err != nil {
		t.Fatal(err)
	}

	args := strings.Split("stack lint", " ")
	args = append(args, "-c", stackFile.Name())

	t.Run("Should parse a valid stack file", func(t *testing.T) {
		ioutil.WriteFile(stackFile.Name(), []byte(valid), commands.File_RW_RW_R)
		defer os.Remove(stackFile.Name())
		err = commands.Run(&LintCmd, args)
		assert.NoError(t, err)
	})

	t.Run("Should fail on parse error", func(t *testing.T) {
		ioutil.WriteFile(stackFile.Name(), []byte(invalidFieldType), commands.File_RW_RW_R)
		defer os.Remove(stackFile.Name())
		err = commands.Run(&LintCmd, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid type")
	})
}
