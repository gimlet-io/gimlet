package release

import (
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimletd/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"os"
)

var releaseTrackCmd = cli.Command{
	Name:  "track",
	Usage: "Track rollback and release requests",
	UsageText: `gimlet release track <id>
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
	},
	Action: track,
}

func track(c *cli.Context) error {
	serverURL := c.String("server")
	token := c.String("token")

	config := new(oauth2.Config)
	auth := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	trackingID := c.Args().First()

	client := client.NewClient(serverURL, auth)
	state, desc, err := client.TrackGet(trackingID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v Request (%s) is %s %s\n", emoji.BackhandIndexPointingRight, trackingID, state, desc)

	return nil
}
