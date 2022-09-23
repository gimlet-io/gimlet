package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		&cli.BoolFlag{
			Name:    "watch",
			Aliases: []string{"w"},
			Usage:   "Updates the output every five seconds",
		},
	},
	Action: track,
}

func track(c *cli.Context) error {
	serverURL := c.String("server")
	token := c.String("token")
	output := c.String("output")
	watch := c.Bool("watch")

	config := new(oauth2.Config)
	auth := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	artifactID := c.Args().First()

	client := client.NewClient(serverURL, auth)

	if watch {
		for {
			err := releaseTrackMessage(client, artifactID, output)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	} else {
		err := releaseTrackMessage(client, artifactID, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func releaseTrackMessage(
	client client.Client,
	artifactID string,
	output string,
) error {
	releaseStatus, err := client.TrackGet(artifactID)
	if err != nil {
		return err
	}

	if output == "json" {
		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(releaseStatus)
		if err != nil {
			return fmt.Errorf("cannot deserialize release status %s", err)
		}

		fmt.Println(jsonString.String())

		return nil
	}

	fmt.Printf(
		"%v Request (%s) is %s %s\n",
		emoji.BackhandIndexPointingRight,
		artifactID,
		releaseStatus.Status,
		releaseStatus.StatusDesc,
	)

	if releaseStatus.Results != nil {
		if len(releaseStatus.Results) == 0 {
			fmt.Printf("\t%v This release don't have any results\n", emoji.Bookmark)
			return nil
		}

		for _, result := range releaseStatus.Results {
			if strings.Contains(result.Status, "fail") {
				fmt.Printf("\t%v App %s on hash %s status is %s, %s\n", emoji.Pager, result.App, result.Hash, result.Status, result.StatusDesc)
			} else {
				fmt.Printf("\t%v App %s on hash %s status is %s\n", emoji.Pager, result.App, result.Hash, result.GitopsCommitStatus)
			}
		}
	} else {
		if len(releaseStatus.GitopsHashes) == 0 {
			fmt.Printf("\t%v This release don't have any gitops hashes\n", emoji.Bookmark)
			return nil
		}

		for _, gitopsHash := range releaseStatus.GitopsHashes {
			if strings.Contains(gitopsHash.Status, "fail") {
				fmt.Printf("\t%v Hash %s status is %s, %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status, gitopsHash.StatusDesc)
			} else {
				fmt.Printf("\t%v Hash %s status is %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status)
			}
		}
	}

	return nil
}
