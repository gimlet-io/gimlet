package artifact

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/commands"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/urfave/cli/v2"
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

const env = `
app: fosdem-2021
env: staging
namespace: default
deploy:
  branch: master
  event: push
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 1
  image:
    repository: ghcr.io/gimlet-io/fosdem-2021
    tag: "{{ .GITHUB_SHA }}"
`

func Test_add(t *testing.T) {
	artifactFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatalf("Error creating artifact file: %s", err)
	}
	defer os.Remove(artifactFile.Name())
	ioutil.WriteFile(artifactFile.Name(), []byte(artifactToExtend), commands.File_RW_RW_R)

	envFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatalf("Error creating env file: %s", err)
	}
	defer os.Remove(envFile.Name())
	ioutil.WriteFile(envFile.Name(), []byte(env), commands.File_RW_RW_R)

	envFile2, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatalf("Error creating env file 2: %s", err)
	}
	defer os.Remove(envFile2.Name())
	ioutil.WriteFile(envFile2.Name(), []byte(env), commands.File_RW_RW_R)

	t.Run("Should add CI URL to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--field", "name=CI")
		args = append(args, "--field", "url=https://jenkins.example.com/job/dev/84/display/redirect")
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error adding CI URL: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(a.Items))
		}

		if a.Items[0]["name"] != "CI" {
			t.Errorf("Expected 'CI', got '%s'", a.Items[0]["name"])
		}

		if a.Items[0]["url"] != "https://jenkins.example.com/job/dev/84/display/redirect" {
			t.Errorf("Expected 'https://jenkins.example.com/job/dev/84/display/redirect', got '%s'", a.Items[0]["url"])
		}

		if len(a.Vars) != 2 {
			t.Errorf("Expected 2 vars, got %d", len(a.Vars))
		}

		if a.Vars["name"] != "CI" {
			t.Errorf("Expected 'CI', got '%s'", a.Vars["name"])
		}

		if a.Vars["url"] != "https://jenkins.example.com/job/dev/84/display/redirect" {
			t.Errorf("Expected 'https://jenkins.example.com/job/dev/84/display/redirect', got '%s'", a.Vars["url"])
		}
	})

	t.Run("Should append custom field to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--field", "custom=myValue")
		app := &cli.App{
			Name: "gimlet",
			Commands: []*cli.Command{
				{
					Name: "artifact",
					Subcommands: []*cli.Command{
						{
							Name: "add",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "file",
									Aliases: []string{"f"},
								},
								&cli.StringSliceFlag{
									Name: "field",
								},
							},
							Action: add,
						},
					},
				},
			},
		}
		if err := app.RunContext(context.TODO(), args); err != nil {
			t.Fatalf("Error appending custom field: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(a.Items))
		}

		if a.Items[1]["custom"] != "myValue" {
			t.Errorf("Expected 'myValue', got '%s'", a.Items[1]["custom"])
		}

		if len(a.Vars) != 3 {
			t.Errorf("Expected 3 vars, got %d", len(a.Vars))
		}

		if a.Vars["custom"] != "myValue" {
			t.Errorf("Expected 'myValue', got '%s'", a.Vars["custom"])
		}
	})

	t.Run("Should add Gimlet environment to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--envFile", envFile.Name())
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Environments) != 1 {
			t.Errorf("Expected 1 env, got %d", len(a.Environments))
		}

		if a.Environments[0].App != "fosdem-2021" {
			t.Errorf("Expected 'fosdem-2021', got '%s'", a.Environments[0].App)
		}
	})

	t.Run("Should append Gimlet environment to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--envFile", envFile2.Name())
		app := &cli.App{
			Name: "gimlet",
			Commands: []*cli.Command{
				{
					Name: "artifact",
					Subcommands: []*cli.Command{
						{
							Name: "add",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "file",
									Aliases: []string{"f"},
								},
								&cli.StringSliceFlag{
									Name: "envFile",
								},
							},
							Action: add,
						},
					},
				},
			},
		}
		if err := app.RunContext(context.TODO(), args); err != nil {
			t.Fatalf("Error: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Environments) != 2 {
			t.Errorf("Expected 2 envs, got %d", len(a.Environments))
		}

		if a.Environments[0].App != "fosdem-2021" {
			t.Errorf("Expected 'fosdem-2021', got '%s'", a.Environments[0].App)
		}
	})

	t.Run("Should add context variables to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--var", "KEY=VALUE")
		args = append(args, "--var", "KEY2=VALUE2")
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Context) != 2 {
			t.Errorf("Expected 2 vars in context, got %d", len(a.Context))
		}

		if a.Context["KEY"] != "VALUE" {
			t.Errorf("Expected 'VALUE', got '%s'", a.Context["KEY"])
		}

		if len(a.Vars) != 5 {
			t.Errorf("Expected 5 variables in vars, got %d", len(a.Vars))
		}

		if a.Vars["KEY"] != "VALUE" {
			t.Errorf("Expected 'VALUE', got '%s'", a.Vars["KEY"])
		}
	})

	t.Run("Should add context variables to artifact that holds an equal sign", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--var", "KEY=https://avatars.githubusercontent.com/u/4289031?v=4")
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if a.Context["KEY"] != "https://avatars.githubusercontent.com/u/4289031?v=4" {
			t.Errorf("Expected 'https://avatars.githubusercontent.com/u/4289031?v=4', got '%s'", a.Context["KEY"])
		}
	})

	t.Run("Should append context variable to artifact", func(t *testing.T) {
		args := strings.Split("gimlet artifact add", " ")
		args = append(args, "-f", artifactFile.Name())
		args = append(args, "--var", "KEY3=VALUE3")
		if err := commands.Run(&Command, args); err != nil {
			t.Fatalf("Error: %s", err)
		}

		content, err := ioutil.ReadFile(artifactFile.Name())
		if err != nil {
			t.Fatalf("Error reading file: %s", err)
		}

		var a dx.Artifact
		if err := json.Unmarshal(content, &a); err != nil {
			t.Fatalf("Error unmarshaling JSON: %s", err)
		}

		if len(a.Context) != 3 {
			t.Errorf("Expected 3 vars in context, got %d", len(a.Context))
		}

		if a.Context["KEY3"] != "VALUE3" {
			t.Errorf("Expected 'VALUE3', got '%s'", a.Context["KEY3"])
		}

		if len(a.Vars) != 6 {
			t.Errorf("Expected 6 variables in vars, got %d", len(a.Vars))
		}

		if a.Vars["KEY3"] != "VALUE3" {
			t.Errorf("Expected 'VALUE3', got '%s'", a.Vars["KEY3"])
		}
	})
}
