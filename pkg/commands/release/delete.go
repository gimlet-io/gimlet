package release

import (
	"context"
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet/pkg/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseDeleteCmd = cli.Command{
	Name:  "delete",
	Usage: "Deletes an application instance",
	UsageText: `gimlet release delete \
     --env staging \
     --app my-app \
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
	},
	Action: delete,
}

func delete(c *cli.Context) error {
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
	err := client.DeletePost(
		c.String("env"),
		c.String("app"),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v Application instance deleted\n", emoji.WomanGesturingOk)

	return nil
}
