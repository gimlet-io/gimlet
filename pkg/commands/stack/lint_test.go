package stack

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
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

func Test_lint(t *testing.T) {
	stackFile, err := ioutil.TempFile("", "stack-test")
	if err != nil {
		t.Fatal(err)
	}

	args := strings.Split("stack lint", " ")
	args = append(args, "-c", stackFile.Name())

	g := goblin.Goblin(t)
	g.Describe("stack lint", func() {
		g.It("Should parse a stack file", func() {
			g.Timeout(time.Second * 10)
			ioutil.WriteFile(stackFile.Name(), []byte(valid), commands.File_RW_RW_R)
			defer os.Remove(stackFile.Name())
			err = commands.Run(&LintCmd, args)
			g.Assert(err == nil).IsTrue(err)
		})
		g.It("Should fail on parse error", func() {
			g.Timeout(time.Second * 10)
			ioutil.WriteFile(stackFile.Name(), []byte(invalidFieldType), commands.File_RW_RW_R)
			defer os.Remove(stackFile.Name())
			err = commands.Run(&LintCmd, args)
			g.Assert(err != nil).IsTrue(err)
			g.Assert(strings.Contains(err.Error(), "Invalid type")).IsTrue(err)
		})
	})
}
