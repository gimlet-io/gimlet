package artifact

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
)

func Test_create(t *testing.T) {
	t.Run("Should create artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact create", " ")

		fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatalf("Error creating temp file: %s", err)
		}
		defer os.Remove(fileToWrite.Name())

		args = append(args, "--repository", "my-app")
		args = append(args, "--sha", "ea9ab7cc31b2599bf4afcfd639da516ca27a4780")
		args = append(args, "--created", "2021-03-19T12:56:03+01:00")
		args = append(args, "--branch", "my-feature")
		args = append(args, "--event", "pr")
		args = append(args, "--sourceBranch", "my-feature")
		args = append(args, "--targetBranch", "main")
		args = append(args, "--authorName", "Jane Doe")
		args = append(args, "--authorEmail", "jane@doe.org")
		args = append(args, "--committerName", "Jane Doe")
		args = append(args, "--committerEmail", "jane@doe.org")
		args = append(args, "--message", "Bugfix 123")
		args = append(args, "--url", "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780")
		args = append(args, "-o", fileToWrite.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatalf("Error running command: %s", err)
		}

		content, err := ioutil.ReadFile(fileToWrite.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		fmt.Println(string(content))
		var a dx.Artifact
		err = json.Unmarshal(content, &a)
		if err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if a.Version.Message != "Bugfix 123" {
			t.Errorf("Expected 'Bugfix 123', got '%s'", a.Version.Message)
		}

		if a.Version.Event != dx.PR {
			t.Errorf("Expected '%s', got '%s'", dx.PR, a.Version.Event)
		}

		if a.Version.Created != 1616154963 {
			t.Errorf("Expected '1616154963', got '%d'", a.Version.Created)
		}
	})
}
