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
			Name:        "timeout",
			Aliases:     []string{"t"},
			Usage:       "Breaks the loop within the given time. Only usable with --wait flag",
			DefaultText: "10m",
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

	artifactID := c.Args().First()

	client := client.NewClient(serverURL, auth)

	if !wait {
		_, err := artifactTrackMessage(client, artifactID, output)
		return err
	}

	timeout := time.After(*timeoutTime)
	for {
		finished, err := artifactTrackMessage(client, artifactID, output)
		if err != nil {
			return err
		}

		if finished {
			return nil
		}

		sleep := time.After(time.Second * 5)
		timedOut := false
		select {
		case <-timeout:
			timedOut = true
		case <-sleep:
		}

		if timedOut {
			break
		}
	}

	return nil
}

func artifactTrackMessage(
	client client.Client,
	artifactID string,
	output string,
) (bool, error) {
	var artifactResultCount int
	var failedCount int
	var succeededCount int
	finished := false

	artifactStatus, err := client.TrackArtifact(artifactID)
	if err != nil {
		return finished, err
	}

	if output == "json" {
		jsonString := bytes.NewBufferString("")
		e := json.NewEncoder(jsonString)
		e.SetIndent("", "  ")
		e.Encode(artifactStatus)
		if err != nil {
			return finished, fmt.Errorf("cannot deserialize release status %s", err)
		}

		fmt.Println(jsonString.String())
		finished = true

		return finished, nil
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

			return finished, nil
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

			return finished, nil
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

	if artifactStatus.Status == "error" || failedCount > 0 {
		err = fmt.Errorf("gitops write failed")
	} else if succeededCount == artifactResultCount && artifactStatus.Status != "new" {
		finished = true
	}

	return finished, err
}
