package release

import (
	"context"
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimletd/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseMakeCmd = cli.Command{
	Name:  "make",
	Usage: "Make an ad-hoc release",
	UsageText: `gimlet release make \
     --env staging \
     --artifact an-artifact-id \
     --server http://gimletd.mycompany.com
     --token c012367f6e6f71de17ae4c6a7baac2e9`,
	Flags: []cli.Flag{
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
		c.String("env"),
		c.String("artifact"),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v Release is now added to the release queue with ID %s\n", emoji.WomanGesturingOk, trackingID)
	fmt.Fprintf(os.Stderr, "Track it with:\ngimlet track %s\n\n", trackingID)

	return nil
}
