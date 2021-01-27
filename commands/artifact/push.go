package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimletd/artifact"
	"github.com/gimlet-io/gimletd/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"io/ioutil"
)

var artifactPushCmd = cli.Command{
	Name:  "push",
	Usage: "Pushes a release artifact to GimletD",
	UsageText: `gimlet artifact push \
     -f artifact.json \
     --server http://gimletd.mycompany.com
     --token c012367f6e6f71de17ae4c6a7baac2e9`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "artifact file to push (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "server",
			Usage:    "GimletD server URL, GIMLET_SERVER environment variable alternatively",
			EnvVars:  []string{"GIMLET_SERVER"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "GimletD server api token, GIMLET_TOKEN environment variable alternatively",
			EnvVars:  []string{"GIMLET_TOKEN"},
			Required: true,
		},
	},
	Action: push,
}

func push(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("cannot read file %s", err)
	}
	var a artifact.Artifact
	err = json.Unmarshal(content, &a)
	if err != nil {
		return fmt.Errorf("cannot parse artifact file %s", err)
	}

	serverURL := c.String("server")
	token := c.String("token")

	config := new(oauth2.Config)
	auth := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	client := client.NewClient(serverURL, auth)

	savedArtifact, err := client.ArtifactPost(&a)
	if err != nil {
		return fmt.Errorf("cannot push artifact file %s", err)
	}

	fmt.Println("Artifact saved")
	savedArtifactStr := bytes.NewBufferString("")
	e := json.NewEncoder(savedArtifactStr)
	e.SetIndent("", "  ")
	err = e.Encode(savedArtifact)
	if err != nil {
		return fmt.Errorf("cannot deserialize artifact %s", err)
	}
	fmt.Println(savedArtifactStr)

	return nil
}
