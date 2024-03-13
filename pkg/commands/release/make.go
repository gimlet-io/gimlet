package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gimlet-io/gimlet/pkg/dx"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet/pkg/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseMakeCmd = cli.Command{
	Name:  "make",
	Usage: "Make an ad-hoc release",
	UsageText: `gimlet release make \
     --env staging \
     --artifact an-artifact-id \
     --server http://gimlet.mycompany.com
     --token c012367f6e6f71de17ae4c6a7baac2e9`,
	Flags: []cli.Flag{
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
			Name:     "env",
			Usage:    "make a release to this environment",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "artifact",
			Usage:    "the artifact to release",
			Aliases:  []string{"a"},
			Required: true,
		},
		&cli.StringFlag{
			Name:  "app",
			Usage: "release only a specific app from the artifact",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Format the output as json with the \"-o json\" switch",
		},
	},
	Action: make,
}

func make(c *cli.Context) error {
	serverURL := c.String("server")
	token := c.String("token")

	config := new(oauth2.Config)
	auth := config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: token,
		},
	)

	client := client.NewClient(serverURL, auth)
	trackingID, err := client.ReleasesPost(
		dx.ReleaseRequest{
			Env:        c.String("env"),
			ArtifactID: c.String("artifact"),
			App:        c.String("app"),
		},
	)
	if err != nil {
		return err
	}

	output := c.String("output")
	if output == "json" {
		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(map[string]interface{}{
			"id": trackingID,
		})
		if err != nil {
			return fmt.Errorf("cannot deserialize json %s", err)
		}

		fmt.Println(jsonString)

		return nil
	}

	fmt.Fprintf(os.Stderr, "%v Release is now added to the release queue with ID %s\n", emoji.WomanGesturingOk, trackingID)
	fmt.Fprintf(os.Stderr, "Track it with:\ngimlet release track %s\n\n", trackingID)

	return nil
}
