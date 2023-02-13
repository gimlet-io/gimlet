package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/artifact"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/rvflash/elapsed"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var releaseListCmd = cli.Command{
	Name:  "list",
	Usage: "Lists releases",
	UsageText: `gimlet release list \
     --app my-app \
     --env staging \
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
			Name:  "app",
			Usage: "filter releases to an application",
		},
		&cli.StringFlag{
			Name:     "env",
			Usage:    "filter envs to an application",
			Required: true,
		},
		&cli.IntFlag{
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
		&cli.BoolFlag{
			Name:    "reverse",
			Aliases: []string{"r"},
			Usage:   "reverse the chronological order of the displayed artifacts",
		},
		&cli.StringFlag{
			Name:  "repo",
			Usage: "filter envs to a source code git repository eg.: laszlocph/myapp",
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

	limit := c.Int("limit")
	if limit == 0 {
		limit = 3
	}

	releases, err := client.ReleasesGet(
		c.String("app"),
		c.String("env"),
		limit,
		c.Int("offset"),
		c.String("repo"),
		since, until,
	)
	if err != nil {
		return err
	}

	if c.String("output") == "json" {
		artifactsStr := bytes.NewBufferString("")
		e := json.NewEncoder(artifactsStr)
		e.SetIndent("", "  ")
		err = e.Encode(releases)
		if err != nil {
			return fmt.Errorf("cannot deserialize releases %s", err)
		}
		fmt.Println(artifactsStr)
	} else {
		if !c.Bool("reverse") { // by default the latest is the bottom of the output
			releases = reverse(releases)
		}
		for _, release := range releases {
			blue := color.New(color.FgBlue, color.Bold).SprintFunc()
			red := color.New(color.FgRed, color.Bold).SprintFunc()
			gray := color.New(color.FgHiBlack).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()

			created := time.Unix(release.Created, 0)

			rolledBack := ""
			if release.RolledBack {
				rolledBack = "**ROLLED BACK**"
			}

			fmt.Printf("%s %s %s %s\n",
				gray(fmt.Sprintf("%s -> %s", release.App, release.Env)),
				blue(fmt.Sprintf("%s@%s", release.GitopsRepo, release.GitopsRef)),
				red(rolledBack),
				green(fmt.Sprintf("(%s)", elapsed.Time(created))),
			)

			if release.Version != nil {
				fmt.Print(artifact.RenderGitVersion(*release.Version, "  "))
			}
			fmt.Println()
		}
	}

	return nil
}

func reverse(input []*dx.Release) []*dx.Release {
	if len(input) == 0 {
		return input
	}
	return append(reverse(input[1:]), input[0])
}
