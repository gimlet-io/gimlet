package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"sigs.k8s.io/yaml"
)

func Test_create(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("gimlet manifest create", func() {
		//g.It("Should resolve <repo>/<chart> format chart names from local helm repo", func() {
		//	createdManifestPath, err := ioutil.TempFile("", "gimlet-cli-test")
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	defer os.Remove(createdManifestPath.Name())
		//
		//	//TODO set up the onechart repo with helm add in this test then this test can run on CI
		//
		//	args := strings.Split("gimlet manifest create", " ")
		//	args = append(args, "--env", "staging")
		//	args = append(args, "--app", "my-app")
		//	args = append(args, "--namespace", "staging")
		//	args = append(args, "--chart", "onechart/onechart")
		//	args = append(args, "-o", createdManifestPath.Name())
		//
		//	err = commands.Run(&Command, args)
		//	g.Assert(err == nil).IsTrue(err)
		//	manifestString, err := ioutil.ReadFile(createdManifestPath.Name())
		//	g.Assert(err == nil).IsTrue(err)
		//
		//	var m dx.Manifest
		//	err = yaml.Unmarshal(manifestString, &m)
		//	g.Assert(err == nil).IsTrue(err)
		//	g.Assert(m.Chart.Repository == "https://chart.onechart.dev").IsTrue("Should resolve chart repo url")
		//})
		g.It("Should resolve git@github.com:gimlet-io/onechart.git format chart names from git", func() {
			createdManifestPath, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(createdManifestPath.Name())

			args := strings.Split("gimlet manifest create", " ")
			args = append(args, "--env", "staging")
			args = append(args, "--app", "my-app")
			args = append(args, "--namespace", "staging")
			args = append(args, "--chart", "git@github.com:gimlet-io/onechart.git?path=/charts/onechart/")
			args = append(args, "-o", createdManifestPath.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)
			manifestString, err := ioutil.ReadFile(createdManifestPath.Name())
			g.Assert(err == nil).IsTrue(err)

			var m dx.Manifest
			err = yaml.Unmarshal(manifestString, &m)
			g.Assert(err == nil).IsTrue(err)
			g.Assert(m.Chart.Name == "git@github.com:gimlet-io/onechart.git?path=/charts/onechart/").IsTrue("Should resolve chart repo url")
		})
	})
}
