package artifact

import (
	"encoding/json"
	"fmt"
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/gimlet-io/gimletd/dx"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_create(t *testing.T) {
	args := strings.Split("gimlet artifact create", " ")
	args = append(args, "--repository", "my-app")
	args = append(args, "--sha", "ea9ab7cc31b2599bf4afcfd639da516ca27a4780")
	args = append(args, "--branch", "master")
	args = append(args, "--event", "pr")
	args = append(args, "--authorName", "Jane Doe")
	args = append(args, "--authorEmail", "jane@doe.org")
	args = append(args, "--committerName", "Jane Doe")
	args = append(args, "--committerEmail", "jane@doe.org")
	args = append(args, "--message", "Bugfix 123")
	args = append(args, "--url", "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780")

	g := goblin.Goblin(t)
	g.Describe("gimlet artifact create", func() {
		g.It("Should create artifact", func() {
			fileToWrite, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(fileToWrite.Name())

			args = append(args, "-o", fileToWrite.Name())
			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			content, err := ioutil.ReadFile(fileToWrite.Name())
			fmt.Println(string(content))
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			g.Assert(err == nil).IsTrue(err)
			g.Assert(a.Version.Message == "Bugfix 123").IsTrue()
			g.Assert(a.Version.Event == dx.PR).IsTrue()
		})
	})
}
