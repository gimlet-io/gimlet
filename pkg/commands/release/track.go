package release

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseTrackCmd = cli.Command{
	Name:  "track",
	Usage: "Track rollback and release requests",
	UsageText: `gimlet release track <id>
     --server http://gimletd.mycompany.com
     --token c012367f6e6f71de17ae4c6a7baac2e9
	 --output json`,
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
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output format",
		},
	},
	Action: track,
}

func track(c *cli.Context) error {
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

	trackingID := c.Args().First()

	client := client.NewClient(serverURL, auth)
	releaseStatus, err := client.TrackGet(trackingID)
	if err != nil {
		return err
	}

	if output == "json" {
		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(releaseStatus)

		fmt.Println(jsonString.String())

		return nil
	}

	fmt.Printf(
		"%v Request (%s) is %s %s\n",
		emoji.BackhandIndexPointingRight,
		trackingID,
		releaseStatus.Status,
		releaseStatus.StatusDesc,
	)

	for _, gitopsHash := range releaseStatus.GitopsHashes {
		fmt.Printf("\t%v Hash %s status is %s\n", emoji.Bookmark, gitopsHash.Hash, gitopsHash.Status)
	}

	for _, result := range releaseStatus.Results {
		fmt.Printf("\t%v App %s on hash %s status is %s\n", emoji.Pager, result.App, result.Hash, result.Status)
	}

	return nil
}
