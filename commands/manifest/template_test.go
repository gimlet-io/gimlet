package manifest

import (
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

const manifestWithRemoteHelmChart = `
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.21.0
values:
  replicas: 1
  image:
    repository: myapp
    tag: 1.1.0
  ingress:
    host: myapp.staging.mycompany.com
    tlsEnabled: true
  volumes:
  - name: uploads
    path: /files
    size: 12Gi
    storageClass: efs-ftp-uploads
  - name: errors
    path: /tmp/err
    size: 12Gi
    storageClass: efs-ftp-errors
`

const manifestWithLocalChart = `
app: myapp
env: staging
namespace: my-team
chart:
  name: ../../fixtures/localChart/hello-server
values:
  replicaCount: 2
`

const manifestWithPrivateGitRepo = `
app: myapp
env: staging
namespace: my-team
chart:
  name: git@github.com:gimlet-io/onechart.git?sha=8e52597ae4fb4ed7888c819b3c77331622136aba&path=/charts/onechart/
values:
  replicas: 10
`

func Test_template(t *testing.T) {
	g := goblin.Goblin(t)

	args := strings.Split("gimlet manifest template", " ")

	g.Describe("gimlet manifest template", func() {
		g.It("Should template a manifest file with remote chart", func() {
			g.Timeout(60*time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithRemoteHelmChart), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "myapp:1.1.0")).IsTrue("Templated manifest should have the image reference")
			//fmt.Println(string(templated))
		})

		g.It("Should template a manifest file with local chart", func() {
			g.Timeout(100*time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithLocalChart), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "hello-server:v0.1.0")).IsTrue("Templated manifest should have the image reference")
			//fmt.Println(string(templated))
		})

		g.It("Should template a manifest file with a private git hosted chart", func() {
			g.Timeout(100*time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithPrivateGitRepo), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "replicas: 10")).IsTrue("should set replicas")
			//fmt.Println(string(templated))
		})
	})
}
