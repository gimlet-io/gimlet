package artifact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimlet-cli/manifest"
	"github.com/gimlet-io/gimletd/artifact"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"io/ioutil"
)

var artifactCreateCmd = cli.Command{
	Name:  "create",
	Usage: "Creates a release artifact",
	UsageText: `gimlet artifact create \
     > artifact.json`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "repository",
			Usage:    "The git repository name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "sha",
			Usage:    "The git sha",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "branch",
			Usage:    "The git branch, or target branch for pull request builds",
			Required: true,
		},
		&cli.BoolFlag{
			Name:     "pr",
			Usage:    "If this is a pull request build",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "sourceBranch",
			Usage:    "For pull requests, the feature branch name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "authorName",
			Usage:    "The person who originally wrote the code",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "authorEmail",
			Usage:    "The person who originally wrote the code",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "committerName",
			Usage:    "The person who originally wrote the code",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "committerEmail",
			Usage:    "The committer is the person who committed the code. Important in case of history rewrite",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "message",
			Usage:    "The git commit message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "url",
			Usage:    "URL to the git commit",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output manifest file",
		},
	},
	Action: create,
}

func create(c *cli.Context) error {
	artifact := &artifact.Artifact{
		ID: c.String("repository") + "-" + uuid.New().String(),
		Version: artifact.Version{
			RepositoryName: c.String("repository"),
			SHA:            c.String("sha"),
			Branch:         c.String("branch"),
			PR:             c.Bool("pr"),
			SourceBranch:   c.String("sourceBranch"),
			AuthorName:     c.String("authorName"),
			AuthorEmail:    c.String("authorEmail"),
			CommitterName:  c.String("committerName"),
			CommitterEmail: c.String("committerEmail"),
			Message:        c.String("message"),
			URL:            c.String("url"),
		},
		Context: map[string]string{

		},
		Environments: []*manifest.Manifest{

		},
		Items: []map[string]interface{}{

		},
	}

	jsonString := bytes.NewBufferString("")
	e := json.NewEncoder(jsonString)
	e.SetIndent("", "  ")
	e.Encode(artifact)

	outputPath := c.String("output")
	if outputPath != "" {
		err := ioutil.WriteFile(outputPath, jsonString.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("cannot write artifact json %s", err)
		}
	} else {
		fmt.Println(jsonString.String())
	}

	return nil
}
