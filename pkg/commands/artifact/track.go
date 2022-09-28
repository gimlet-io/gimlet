package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var artifactTrackCmd = cli.Command{
	Name:  "track",
	Usage: "Track artifact from release requests",
	UsageText: `gimlet artifact track <artifact_id>
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

	if timeoutString != "" && !wait {
		return fmt.Errorf("--wait flag is required with --timeout")
	}

	config := new(oauth2.Config)
	auth := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	artifactID := c.Args().First()

	client := client.NewClient(serverURL, auth)

	if wait {
		if timeoutString != "" {
			timeoutTime, err := time.ParseDuration(timeoutString)
			if err != nil {
				return err
			}

		loop:
			for timeout := time.After(timeoutTime); ; {
				select {
				case <-timeout:
					break loop
				default:
				}
				artifactStatus, hasFailed, everySucceeded, err := artifactTrackMessage(client, artifactID, output)
				if err != nil {
					return err
				}
				if (artifactStatus == "error" || hasFailed || everySucceeded) && artifactStatus != "new" {
					break
				}
				time.Sleep(time.Second * 5)
			}
		} else {
			for {
				artifactStatus, hasFailed, everySucceeded, err := artifactTrackMessage(client, artifactID, output)
				if err != nil {
					return err
				}
				if (artifactStatus == "error" || hasFailed || everySucceeded) && artifactStatus != "new" {
					break
				}
				time.Sleep(time.Second * 5)
			}
		}
	} else {
		_, _, _, err := artifactTrackMessage(client, artifactID, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func artifactTrackMessage(
	client client.Client,
	artifactID string,
	output string,
) (string, bool, bool, error) {
	var artifactResultCount int
	var failedCount int
	var succeededCount int

	artifactStatus, err := client.TrackArtifact(artifactID)
	if err != nil {
		return "", false, false, err
	}

	if output == "json" {
		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(artifactStatus)
		if err != nil {
			return "", false, false, fmt.Errorf("cannot deserialize release status %s", err)
		}

		fmt.Println(jsonString.String())

		return "", false, true, nil
	}

	fmt.Printf(
		"%v Request (%s) is %s %s\n",
		emoji.BackhandIndexPointingRight,
		artifactID,
		artifactStatus.Status,
		artifactStatus.StatusDesc,
	)

	if artifactStatus.Results != nil {
		if len(artifactStatus.Results) == 0 {
			fmt.Printf("\t%v This release don't have any results\n", emoji.Bookmark)

			return "", false, false, nil
		}

		artifactResultCount = len(artifactStatus.Results)

		for _, result := range artifactStatus.Results {
			if strings.Contains(result.GitopsCommitStatus, "Failed") {
				failedCount++
				fmt.Printf("\t%v App %s on %s hash %s status is %s, %s\n", emoji.Pager, result.App, result.Env, result.Hash, result.Status, result.StatusDesc)
			} else {
				if result.GitopsCommitStatus == model.ReconciliationSucceeded {
					succeededCount++
				}

				fmt.Printf("\t%v App %s on %s hash %s status is %s\n", emoji.Pager, result.App, result.Env, result.Hash, result.GitopsCommitStatus)
			}
		}
	} else {
		if len(artifactStatus.GitopsHashes) == 0 {
			fmt.Printf("\t%v This release don't have any gitops hashes\n", emoji.Bookmark)

			return "", false, false, nil
		}

		artifactResultCount = len(artifactStatus.GitopsHashes)

		for _, gitopsHash := range artifactStatus.GitopsHashes {
			if strings.Contains(gitopsHash.Status, "Failed") {
				failedCount++
				fmt.Printf("\t%v Hash %s status is %s, %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status, gitopsHash.StatusDesc)
			} else {
				if gitopsHash.Status == model.ReconciliationSucceeded {
					succeededCount++
				}
				fmt.Printf("\t%v Hash %s status is %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status)
			}
		}
	}

	return artifactStatus.Status, failedCount > 0, succeededCount == artifactResultCount, nil
}
