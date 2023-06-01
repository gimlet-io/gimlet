package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	// "github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
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
		t.Fatal(err)
	}
	defer os.Remove(artifactFile.Name())
	ioutil.WriteFile(artifactFile.Name(), []byte(artifactToExtend), commands.File_RW_RW_R)

	envFile, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(envFile.Name())
	ioutil.WriteFile(envFile.Name(), []byte(env), commands.File_RW_RW_R)

	envFile2, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(envFile2.Name())
	ioutil.WriteFile(envFile2.Name(), []byte(env), commands.File_RW_RW_R)

	t.Run("gimlet artifact add", func(t *testing.T) {
		t.Run("Should add CI URL to artifact", func(t *testing.T) {
			args := strings.Split("gimlet artifact add", " ")
			args = append(args, "-f", artifactFile.Name())
			args = append(args, "--field", "name=CI")
			args = append(args, "--field", "url=https://jenkins.example.com/job/dev/84/display/redirect")
			err := commands.Run(&Command, args)
			assert.NoError(t, err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Items, 1)
			assert.Equal(t, "CI", a.Items[0]["name"])
			assert.Equal(t, "https://jenkins.example.com/job/dev/84/display/redirect", a.Items[0]["url"])
			assert.Len(t, a.Vars, 2)
			assert.Equal(t, "CI", a.Vars["name"])
			assert.Equal(t, "https://jenkins.example.com/job/dev/84/display/redirect", a.Vars["url"])
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
			err = app.RunContext(context.TODO(), args)
			assert.NoError(t, err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Items, 2)
			assert.Equal(t, "myValue", a.Items[1]["custom"])
			assert.Len(t, a.Vars, 3)
			assert.Equal(t, "myValue", a.Vars["custom"])
		})

		t.Run("Should add Gimlet environment to artifact", func(t *testing.T) {
			args := strings.Split("gimlet artifact add", " ")
			args = append(args, "-f", artifactFile.Name())
			args = append(args, "--envFile", envFile.Name())
			err = commands.Run(&Command, args)
			assert.NoError(t, err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Environments, 1)
			assert.Equal(t, "fosdem-2021", a.Environments[0].App)
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
			err = app.RunContext(context.TODO(), args)
			assert.NoError(t, err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Environments, 2)
			assert.Equal(t, "fosdem-2021", a.Environments[0].App)
			fmt.Println(string(content))
		})

		t.Run("Should add context variables to artifact", func(t *testing.T) {
			args := strings.Split("gimlet artifact add", " ")
			args = append(args, "-f", artifactFile.Name())
			args = append(args, "--var", "KEY=VALUE")
			args = append(args, "--var", "KEY2=VALUE2")
			err = commands.Run(&Command, args)
			assert.NoError(t, err)

			content, err := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Context, 2)
			assert.Equal(t, "VALUE", a.Context["KEY"])
			assert.Len(t, a.Vars, 5)
			assert.Equal(t, "VALUE", a.Vars["KEY"])
		})

		t.Run("Should add context variables to artifact that holds an equal sign", func(t *testing.T) {
			args := strings.Split("gimlet artifact add", " ")
			args = append(args, "-f", artifactFile.Name())
			args = append(args, "--var", "KEY=https://avatars.githubusercontent.com/u/4289031?v=4")
			err = commands.Run(&Command, args)
			assert.NoError(t, err)

			content, _ := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Equal(t, "https://avatars.githubusercontent.com/u/4289031?v=4", a.Context["KEY"])
		})
		t.Run("Should append context variable to artifact", func(t *testing.T) {
			args := strings.Split("gimlet artifact add", " ")
			args = append(args, "-f", artifactFile.Name())
			args = append(args, "--var", "KEY3=VALUE3")
			err = commands.Run(&Command, args)
			assert.NoError(t, err)

			content, _ := ioutil.ReadFile(artifactFile.Name())
			var a dx.Artifact
			err = json.Unmarshal(content, &a)
			assert.NoError(t, err)
			assert.Len(t, a.Context, 3)
			assert.Equal(t, "VALUE3", a.Context["KEY3"])
			assert.Len(t, a.Vars, 6)
			assert.Equal(t, "VALUE3", a.Vars["KEY3"])
		})
	})
}


	
