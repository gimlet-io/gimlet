package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
)

const githubCommitLink = "https://github.com/%s/commit/%s"
const contextFormat = "gitops/%s@%s"

type gitopsDeployMessage struct {
	event model.Result
}

func (gm *gitopsDeployMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	if gm.event.Status == model.Failure {
		msg.Text = fmt.Sprintf("Failed to roll out %s of %s", gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: msg.Text,
				},
			},
		)
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: contextString,
				Elements: []Text{
					{
						Type: markdown,
						Text: fmt.Sprintf(":exclamation: *Error* :exclamation: \n%s", gm.event.StatusDesc),
					},
				},
			},
		)
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: contextString,
				Elements: []Text{
					{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.event.Manifest.Env))},
					{Type: markdown, Text: fmt.Sprintf(":clipboard: %s", gm.event.Artifact.Version.URL)},
				},
			},
		)
	} else {
		if gm.event.TriggeredBy == "policy" {
			msg.Text = fmt.Sprintf("Policy based rollout of %s on %s", gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)
		} else {
			msg.Text = fmt.Sprintf("%s is rolling out %s on %s", gm.event.TriggeredBy, gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)
		}
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: msg.Text,
				},
			},
		)
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: contextString,
				Elements: []Text{
					{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.event.Manifest.Env))},
					{Type: markdown, Text: fmt.Sprintf(":clipboard: %s", gm.event.Artifact.Version.URL)},
					{Type: markdown, Text: fmt.Sprintf(":paperclip: %s", commitLink(gm.event.GitopsRepo, gm.event.GitopsRef))},
				},
			},
		)
	}

	return msg, nil
}

func (gm *gitopsDeployMessage) Env() string {
	return gm.event.Manifest.Env
}

func (gm *gitopsDeployMessage) AsStatus() (*status, error) {
	context := fmt.Sprintf(contextFormat, gm.event.Manifest.Env, time.Now().Format(time.RFC3339))
	desc := gm.event.StatusDesc
	if len(desc) > 140 {
		desc = desc[:140]
	}

	state := "success"
	targetURL := fmt.Sprintf(githubCommitLink, gm.event.GitopsRepo, gm.event.GitopsRef)

	if gm.event.Status == model.Failure {
		state = "failure"
		targetURL = ""
	}

	return &status{
		state:       state,
		context:     context,
		description: desc,
		targetURL:   targetURL,
	}, nil
}

func (gm *gitopsDeployMessage) AsDiscordMessage() (*discordMessage, error) {

	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	if gm.event.Status == model.Failure {
		msg.Text = fmt.Sprintf("Failed to roll out %s of %s", gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)

		msg.Embed.Description += fmt.Sprintf(":exclamation: *Error* :exclamation: \n%s\n", gm.event.StatusDesc)
		msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.event.Manifest.Env))
		msg.Embed.Description += fmt.Sprintf(":clipboard: %s\n", gm.event.Artifact.Version.URL)

		msg.Embed.Color = 15158332

	} else {
		if gm.event.TriggeredBy == "policy" {
			msg.Text = fmt.Sprintf("Policy based rollout of %s on %s", gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)
		} else {
			msg.Text = fmt.Sprintf("%s is rolling out %s on %s", gm.event.TriggeredBy, gm.event.Manifest.App, gm.event.Artifact.Version.RepositoryName)
		}

		msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.event.Manifest.Env))
		msg.Embed.Description += fmt.Sprintf(":clipboard: %s\n", gm.event.Artifact.Version.URL)
		msg.Embed.Description += fmt.Sprintf(":paperclip: %s\n", discordCommitLink(gm.event.GitopsRepo, gm.event.GitopsRef))

		msg.Embed.Color = 3066993

	}

	return msg, nil

}

func DeployMessageFromGitOpsResult(event model.Result) Message {
	return &gitopsDeployMessage{
		event: event,
	}
}

func (gm *gitopsDeployMessage) RepositoryName() string {
	return gm.event.Artifact.Version.RepositoryName
}

func (gm *gitopsDeployMessage) SHA() string {
	return gm.event.Artifact.Version.SHA
}
