package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
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
			Usage:   "Format the output as json with the \"-o json\" switch",
		},
		&cli.BoolFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "Wait until the artifact is processed",
		},
		&cli.StringFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Usage:   "If you specified the wait flag, the wait will time out by this specified value. The default is 10m (minutes)",
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

		if releaseStatus.Status == model.StatusNew {
			fmt.Printf("\t%v The release is not processed yet...\n", emoji.HourglassNotDone)
		} else if releaseStatus.Status == model.StatusError {
			return fmt.Errorf(releaseStatus.StatusDesc)
		} else {
			printGitopsStatuses(releaseStatus)
			allGitopsCommitsApplied, gitopsCommitsHaveFailed := releaseStatus.ExtractGitopsEndState()
			if gitopsCommitsHaveFailed {
				return fmt.Errorf("gitops commits have failed to apply")
			} else if allGitopsCommitsApplied {
				return nil
			}
		}

		if !wait {
			break
		}

		sleep := time.After(time.Second * 5)
		select {
		case <-timeout:
			return fmt.Errorf("process timed out")
		case <-sleep:
		}
	}

	return nil
}

func printGitopsStatuses(releaseStatus *dx.ReleaseStatus) {
	if releaseStatus.Results != nil {
		if len(releaseStatus.Results) == 0 {
			fmt.Printf("\t%v The release didn't generate any gitops commits\n", emoji.Bookmark)
		}

		for _, result := range releaseStatus.Results {
			if result.Status == model.Failure.String() {
				fmt.Printf("\t%v %s -> %s, status is %s, %s\n", emoji.ExclamationMark, result.App, result.Env, result.Status, result.StatusDesc)
			} else if strings.Contains(result.GitopsCommitStatus, "Failed") {
				fmt.Printf("\t%v %s -> %s, gitops hash %s, status is %s, %s\n", emoji.ExclamationMark, result.App, result.Env, result.Hash, result.Status, result.StatusDesc)
			} else {
				fmt.Printf("\t%v %s -> %s, gitops hash %s, status is %s\n", emoji.OpenBook, result.App, result.Env, result.Hash, result.GitopsCommitStatus)
			}
		}
	} else {
		if len(releaseStatus.GitopsHashes) == 0 {
			fmt.Printf("\t%v The release didn't generate any gitops commits\n", emoji.Bookmark)
		}

		for _, gitopsHash := range releaseStatus.GitopsHashes {
			if releaseStatus.Status == model.Failure.String() {
				fmt.Printf("\t%v status is %s, %s\n", emoji.ExclamationMark, releaseStatus.Status, releaseStatus.StatusDesc)
			} else if strings.Contains(gitopsHash.Status, "Failed") {
				fmt.Printf("\t%v Gitops hash %s status is %s, %s\n", emoji.ExclamationMark, gitopsHash.Hash, gitopsHash.Status, gitopsHash.StatusDesc)
			} else {
				fmt.Printf("\t%v Gitops hash %s status is %s\n", emoji.OpenBook, gitopsHash.Hash, gitopsHash.Status)
			}
		}
	}
}
