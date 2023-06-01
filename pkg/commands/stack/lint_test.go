package stack

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	// "github.com/franela/goblin"
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

func TestLint(t *testing.T) {
	stackFile, err := ioutil.TempFile("", "stack-test")
	if err != nil {
		t.Fatal(err)
	}

	args := strings.Split("stack lint", " ")
	args = append(args, "-c", stackFile.Name())

	t.Run("Should parse a stack file", func(t *testing.T) {
		t.Timeout(time.Second * 10)
		valid := "valid stack content"
		err = ioutil.WriteFile(stackFile.Name(), []byte(valid), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(stackFile.Name())

		// Call your lint function here
		err = commands.Run(&LintCmd, args)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("Should fail on parse error", func(t *testing.T) {
		t.Timeout(time.Second * 10)
		invalidFieldType := "invalid stack content"
		err = ioutil.WriteFile(stackFile.Name(), []byte(invalidFieldType), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(stackFile.Name())

		// Call your lint function here
		err = commands.Run(&LintCmd, args)
		if err == nil {
			t.Error("Expected an error, but got nil")
		} else {
			expectedErrMsg := "Invalid type"
			if !strings.Contains(err.Error(), expectedErrMsg) {
				t.Errorf("Expected error message to contain '%s', but got: %v", expectedErrMsg, err)
			}
		}
	})
}

		
}
