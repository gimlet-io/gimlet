package release

import (
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimletd/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"os"
)

var releaseRollbackCmd = cli.Command{
	Name:  "rollback",
	Usage: "Rolls back to the desired sha",
	UsageText: `gimlet release rollback \
     --env staging \
     --app my-app \
     --to a-release-sha \
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
		oauth2.NoContext,
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

	fmt.Fprintf(os.Stderr, "%v Your rollback request is now added to the release queue with ID %s\n\n", emoji.WomanGesturingOk, trackingID)

	return nil
}
