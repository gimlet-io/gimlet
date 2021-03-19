package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gimlet-io/gimletd/client"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/rvflash/elapsed"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"strings"
	"time"
)

var artifactListCmd = cli.Command{
	Name:  "list",
	Usage: "Lists the releasable artifacts",
	UsageText: `gimlet artifact list \
     --app my-app \
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
			Name:  "app",
			Usage: "filter artifacts to an application",
		},
		&cli.StringFlag{
			Name:  "branch",
			Usage: "filter artifacts to a branch",
		},
		&cli.StringFlag{
			Name:  "event",
			Usage: "filter artifacts to a git event",
		},
		&cli.StringFlag{
			Name:  "sourceBranch",
			Usage: "filter PR artifacts to a source branch",
		}, &cli.StringFlag{
			Name:  "sha",
			Usage: "filter artifacts to a git SHA",
		}, &cli.IntFlag{
			Name:  "limit",
			Usage: "limit the number of returned artifacts",
		}, &cli.IntFlag{
			Name:  "offset",
			Usage: "offset the returned artifacts",
		}, &cli.StringFlag{
			Name:  "since",
			Usage: "the RFC3339 format date to return the artifacts from (eg 2021-02-01T15:34:26+01:00)",
		}, &cli.StringFlag{
			Name:  "until",
			Usage: "the RFC3339 format date to return the artifacts until (eg 2021-02-01T15:34:26+01:00)",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output format, eg.: json",
		},
	},
	Action: list,
}

func list(c *cli.Context) error {
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

	var since, until *time.Time
	var err error
	if c.String("since") != "" {
		t, err := time.Parse(time.RFC3339, c.String("since"))
		if err != nil {
			return fmt.Errorf("cannot parse since date %s", err)
		}
		since = &t
	}
	if c.String("until") != "" {
		t, err := time.Parse(time.RFC3339, c.String("until"))
		if err != nil {
			return fmt.Errorf("cannot parse until date %s", err)
		}
		until = &t
	}

	var event *dx.GitEvent
	if c.String("event") != "" {
		event = dx.PushPtr()
		err := event.UnmarshalJSON([]byte(`"` + c.String("event") + `"`))
		if err != nil {
			return fmt.Errorf("cannot parse event: %s", err)
		}
	}

	artifacts, err := client.ArtifactsGet(
		c.String("app"), c.String("branch"),
		event,
		c.String("sourceBranch"),
		c.String("sha"),
		c.Int("limit"), c.Int("offset"),
		since, until,
	)

	if err != nil {
		return err
	}

	if c.String("output") == "json" {
		artifactsStr := bytes.NewBufferString("")
		e := json.NewEncoder(artifactsStr)
		e.SetIndent("", "  ")
		err = e.Encode(artifacts)
		if err != nil {
			return fmt.Errorf("cannot deserialize artifacts %s", err)
		}
		fmt.Println(artifactsStr)
	} else {
		for _, artifact := range artifacts {
			blue := color.New(color.FgBlue, color.Bold).SprintFunc()
			gray := color.New(color.FgHiBlack).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()

			created := time.Unix(artifact.Created, 0)

			fmt.Printf("%s\n", yellow(artifact.ID))
			fmt.Printf("%s - %s %s %s\n", red(artifact.Version.SHA[:8]), limitMessage(makeSingleLine(artifact.Version.Message)), green(fmt.Sprintf("(%s)", elapsed.Time(created))), blue(artifact.Version.CommitterName))
			fmt.Printf("%s %s\n", artifact.Version.RepositoryName, green(artifact.Version.Branch))
			fmt.Println(gray(artifact.Version.URL))
			fmt.Println()
		}
	}

	return nil
}

func makeSingleLine(message string) string {
	message = strings.ReplaceAll(message, "\n\n", "\n")
	message = strings.ReplaceAll(message, "\n", "; ")
	return message
}

func limitMessage(message string) string{
	if len(message) > 80 {
		message = message[0:79]
		message = message + "..."
	}

	return message
}
