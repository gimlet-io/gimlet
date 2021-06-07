package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gimlet-io/gimlet-cli/commands/artifact"
	"github.com/gimlet-io/gimletd/client"
	"github.com/rvflash/elapsed"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"sort"
	"time"
)

var releaseStatusCmd = cli.Command{
	Name:  "status",
	Usage: "Lists apps and current release",
	UsageText: `gimlet release status \
     --env staging \
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
			Usage: "filter releases to an application",
		},
		&cli.StringFlag{
			Name:     "env",
			Usage:    "filter envs to an application",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output format, eg.: json",
		},
	},
	Action: status,
}

func status(c *cli.Context) error {
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

	var err error

	appReleases, err := client.StatusGet(
		c.String("app"),
		c.String("env"),
	)
	if err != nil {
		return err
	}

	if c.String("output") == "json" {
		appReleasesString := bytes.NewBufferString("")
		e := json.NewEncoder(appReleasesString)
		e.SetIndent("", "  ")
		err = e.Encode(appReleasesString)
		if err != nil {
			return fmt.Errorf("cannot deserialize appReleases %s", err)
		}
		fmt.Println(appReleasesString)
	} else {
		var keys []string
		for k := range appReleases {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, app := range keys {
			release := appReleases[app]

			blue := color.New(color.FgBlue, color.Bold).SprintFunc()
			red := color.New(color.FgRed, color.Bold).SprintFunc()
			gray := color.New(color.FgHiBlack).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()

			fmt.Println(app)

			if release != nil {
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
			} else {
				fmt.Println(gray("Release data not available"))
			}

			fmt.Println()
		}
	}

	return nil
}
