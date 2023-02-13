package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var artifactPushCmd = cli.Command{
	Name:  "push",
	Usage: "Pushes a release artifact to Gimlet",
	UsageText: `gimlet artifact push \
     -f artifact.json \
     --server http://gimlet.mycompany.com
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
			Usage:    "Gimlet server URL, GIMLET_SERVER environment variable alternatively",
			EnvVars:  []string{"GIMLET_SERVER"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "Gimlet server api token, GIMLET_TOKEN environment variable alternatively",
			EnvVars:  []string{"GIMLET_TOKEN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output format",
		},
	},
	Action: push,
}

func push(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("cannot read file %s", err)
	}
	var a dx.Artifact
	err = json.Unmarshal(content, &a)
	if err != nil {
		return fmt.Errorf("cannot parse artifact file %s", err)
	}

	serverURL := c.String("server")
	token := c.String("token")
	output := c.String("output")

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

	savedArtifactStr := bytes.NewBufferString("")
	e := json.NewEncoder(savedArtifactStr)
	e.SetIndent("", "  ")
	err = e.Encode(savedArtifact)
	if err != nil {
		return fmt.Errorf("cannot deserialize artifact %s", err)
	}

	if output == "json" {
		fmt.Println(savedArtifactStr)

		return nil
	}

	fmt.Println("Artifact saved")
	fmt.Println(savedArtifactStr)

	return nil
}
