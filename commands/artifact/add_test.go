package artifact

import (
	"encoding/json"
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/gimlet-io/gimletd/artifact"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const artifactToExtend = `
{
  "id": "my-app-b2ab0f7a-ca0e-45cf-83a0-cadd94dddeac",
  "version": {
    "sha": "ea9ab7cc31b2599bf4afcfd639da516ca27a4780",
    "branch": "master",
    "authorName": "Jane Doe",
    "authorEmail": "jane@doe.org",
    "committerName": "Jane Doe",
    "committerEmail": "jane@doe.org",
    "message": "Bugfix 123",
    "repositoryName": "my-app",
    "url": "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780"
  }
}
`

func Test_add(t *testing.T) {
	artifactFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(artifactFile.Name())

	ioutil.WriteFile(artifactFile.Name(), []byte(artifactToExtend), commands.File_RW_RW_R)

	args := strings.Split("gimlet artifact add", " ")
	args = append(args, "-f", artifactFile.Name())
	args = append(args, "--field", "name=CI")
	args = append(args, "--field", "url=https://jenkins.example.com/job/dev/84/display/redirect")

	g := goblin.Goblin(t)
	g.Describe("gimlet artifact add", func() {
		g.It("Should add CI URL to artifact", func() {
			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a artifact.Artifact
			err = json.Unmarshal(content, &a)
			g.Assert(err == nil).IsTrue(err)
			g.Assert(len(a.Items) == 1).IsTrue("Should have 1 item")
			g.Assert(a.Items[0]["name"] == "CI").IsTrue("Should add CI item")
			g.Assert(a.Items[0]["url"] == "https://jenkins.example.com/job/dev/84/display/redirect").IsTrue("Should add CI item")
			//fmt.Println(string(content))
		})
	})
}