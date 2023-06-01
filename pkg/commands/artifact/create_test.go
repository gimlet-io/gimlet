package artifact

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	// "github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
)

func Test_create(t *testing.T) {
	args := []string{"gimlet", "artifact", "create"}
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

	t.Run("Should create artifact", func(t *testing.T) {
		fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(fileToWrite.Name())

		args = append(args, "-o", fileToWrite.Name())
		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatalf("Error running command: %s", err)
		}

		content, err := ioutil.ReadFile(fileToWrite.Name())
		fmt.Println(string(content))
		var a dx.Artifact
		err = json.Unmarshal(content, &a)
		if err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		expectedMessage := "Bugfix 123"
		if a.Version.Message != expectedMessage {
			t.Errorf("Expected message to be %s, but got %s", expectedMessage, a.Version.Message)
		}

		expectedEvent := dx.PR
		if a.Version.Event != expectedEvent {
			t.Errorf("Expected event to be %s, but got %s", expectedEvent, a.Version.Event)
		}

		expectedCreated := time.Unix(1616154963, 0)
		if a.Version.Created != expectedCreated {
			t.Errorf("Expected created time to be %v, but got %v", expectedCreated, a.Version.Created)
		}
	})
}
