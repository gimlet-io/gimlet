package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/commands"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"sigs.k8s.io/yaml"
)

func Test_create(t *testing.T) {
	t.Run("Should resolve <repo>/<chart> format chart names from local helm repo", func(t *testing.T) {
		createdManifestPath, err := ioutil.TempFile("", "gimlet-cli-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(createdManifestPath.Name())

		args := strings.Split("gimlet manifest create", " ")
		args = append(args, "--env", "staging")
		args = append(args, "--app", "my-app")
		args = append(args, "--namespace", "staging")
		args = append(args, "--chart", "onechart/onechart")
		args = append(args, "-o", createdManifestPath.Name())

		err = commands.Run(&Command, args)
		if err != nil {
			t.Fatal(err)
		}
		manifestString, err := ioutil.ReadFile(createdManifestPath.Name())
		if err != nil {
			t.Fatal(err)
		}

		var m dx.Manifest
		err = yaml.Unmarshal(manifestString, &m)
		if err != nil {
			t.Fatal(err)
		}
		if m.Chart.Repository != "https://chart.onechart.dev" {
			t.Error("Should resolve chart repo url")
		}
	})

	t.Run("Should resolve git@github.com:gimlet-io/onechart.git format chart names from git", func(t *testing.T) {
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
		if err != nil {
			t.Fatal(err)
		}
		manifestString, err := ioutil.ReadFile(createdManifestPath.Name())
		if err != nil {
			t.Fatal(err)
		}

		var m dx.Manifest
		err = yaml.Unmarshal(manifestString, &m)
		if err != nil {
			t.Fatal(err)
		}
		if m.Chart.Name != "git@github.com:gimlet-io/onechart.git?path=/charts/onechart/" {
			t.Error("Should resolve chart repo url")
		}
	})
}
