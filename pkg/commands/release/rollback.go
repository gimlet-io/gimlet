package release

import (
	"context"
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseRollbackCmd = cli.Command{
	Name:  "rollback",
	Usage: "Rolls back to the desired sha",
	UsageText: `gimlet release rollback \
     --env staging \
     --app my-app \
     --to a-release-sha \
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
			Usage:    "rollback in this environment",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "app",
			Usage:    "rollback this app",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "to",
			Usage:    "rollback to this sha",
			Aliases:  []string{"t"},
			Required: true,
		},
	},
	Action: rollback,
}

func rollback(c *cli.Context) error {
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
	trackingID, err := client.RollbackPost(
		c.String("env"),
		c.String("app"),
		c.String("to"),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v Your rollback request is now added to the release queue with ID %s\n", emoji.WomanGesturingOk, trackingID)
	fmt.Fprintf(os.Stderr, "Track it with:\ngimlet release track %s\n\n", trackingID)

	return nil
}
