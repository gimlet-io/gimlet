package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/artifact"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
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
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output format. Cannot use with --wait flag",
		},
		&cli.BoolFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "Updates the output every five seconds. Runs until Artifact has error, at least one gitops hash has error or every gitops has has succeeded. Cannot use with --output flag",
		},
		&cli.StringFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Usage:   "Breaks the loop within the given time. Only usable with --wait flag",
			Value:   "10m",
		},
	},
	Action: track,
}

func track(c *cli.Context) error {
	serverURL := c.String("server")
	token := c.String("token")
	output := c.String("output")
	wait := c.Bool("wait")
	timeoutString := c.String("timeout")

	var timeoutTime *time.Duration
	t, err := time.ParseDuration(timeoutString)
	if err != nil {
		return err
	}
	timeoutTime = &t

	config := new(oauth2.Config)
	auth := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	trackingID := c.Args().First()

	client := client.NewClient(serverURL, auth)

	if output == "json" {
		releaseStatus, err := client.TrackRelease(trackingID)
		if err != nil {
			return err
		}

		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(releaseStatus)
		if err != nil {
			return fmt.Errorf("cannot deserialize release status %s", err)
		}

		fmt.Println(jsonString)

		return nil
	}

	var gitopsCommitsStatusDesc strings.Builder
	timeout := time.After(*timeoutTime)
	for {
		releaseStatus, err := client.TrackRelease(trackingID)
		if err != nil {
			return err
		}

		fmt.Printf(
			"%v Request (%s) is %s %s\n",
			emoji.BackhandIndexPointingRight,
			trackingID,
			releaseStatus.Status,
			releaseStatus.StatusDesc,
		)

		if releaseStatus.Results != nil {
			if len(releaseStatus.Results) == 0 {
				fmt.Printf("\t%v This release don't have any results\n", emoji.Bookmark)
			}

			for _, result := range releaseStatus.Results {
				if strings.Contains(result.GitopsCommitStatus, "Failed") {
					gitopsCommitsStatusDesc.WriteString(result.StatusDesc + "\n")
					fmt.Printf("\t%v App %s on %s hash %s status is %s, %s\n", emoji.Pager, result.App, result.Env, result.Hash, result.Status, result.StatusDesc)
				} else {
					fmt.Printf("\t%v App %s on %s hash %s status is %s\n", emoji.Pager, result.App, result.Env, result.Hash, result.GitopsCommitStatus)
				}
			}
		} else {
			if len(releaseStatus.GitopsHashes) == 0 {
				fmt.Printf("\t%v This release don't have any gitops hashes\n", emoji.Bookmark)
			}

			for _, gitopsHash := range releaseStatus.GitopsHashes {
				if strings.Contains(gitopsHash.Status, "Failed") {
					gitopsCommitsStatusDesc.WriteString(gitopsHash.StatusDesc + "\n")
					fmt.Printf("\t%v Hash %s status is %s, %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status, gitopsHash.StatusDesc)
				} else {
					fmt.Printf("\t%v Hash %s status is %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status)
				}
			}
		}

		artifactProcessingError, everythingSucceeded, gitopsCommitsHaveFailed := artifact.ExtractEndState(*releaseStatus)

		if artifactProcessingError {
			return fmt.Errorf(releaseStatus.StatusDesc)
		} else if gitopsCommitsHaveFailed {
			return fmt.Errorf(gitopsCommitsStatusDesc.String())
		} else if everythingSucceeded {
			return nil
		}

		sleep := time.After(time.Second * 5)
		timedOut := false
		select {
		case <-timeout:
			timedOut = true
		case <-sleep:
		}

		if timedOut || !wait {
			break
		}
	}

	return nil
}
