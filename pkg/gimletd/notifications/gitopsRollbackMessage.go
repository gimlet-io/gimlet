package notifications

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker/events"
	githubLib "github.com/google/go-github/v37/github"
)

type gitopsRollbackMessage struct {
	event *events.RollbackEvent
}

func (gm *gitopsRollbackMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	if gm.event.Status == events.Failure {
		msg.Text = fmt.Sprintf("Failed to roll back %s of %s",
			gm.event.RollbackRequest.App,
			gm.event.RollbackRequest.Env)
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
					{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.event.RollbackRequest.Env))},
					{Type: markdown, Text: fmt.Sprintf(":clipboard: %s", gm.event.RollbackRequest.TargetSHA)},
				},
			},
		)
	} else {
		msg.Text = fmt.Sprintf("🔙 %s is rolling back %s on %s", gm.event.RollbackRequest.TriggeredBy, gm.event.RollbackRequest.App, gm.event.RollbackRequest.Env)
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
					{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.event.RollbackRequest.Env))},
					{Type: markdown, Text: fmt.Sprintf(":clipboard: %s", gm.event.RollbackRequest.TargetSHA)},
				},
			},
		)
		for _, gitopsRef := range gm.event.GitopsRefs {
			msg.Blocks[len(msg.Blocks)-1].Elements = append(
				msg.Blocks[len(msg.Blocks)-1].Elements,
				Text{Type: markdown, Text: fmt.Sprintf(":paperclip: %s", commitLink(gm.event.GitopsRepo, gitopsRef))},
			)
		}
		if len(msg.Blocks[len(msg.Blocks)-1].Elements) > 10 {
			msg.Blocks[len(msg.Blocks)-1].Elements = msg.Blocks[len(msg.Blocks)-1].Elements[:10]
		}
	}

	return msg, nil
}

func (gm *gitopsRollbackMessage) Env() string {
	return gm.event.RollbackRequest.Env
}

func (gm *gitopsRollbackMessage) AsGithubStatus() (*githubLib.RepoStatus, error) {
	return nil, nil
}

func (gm *gitopsRollbackMessage) AsDiscordMessage() (*discordMessage, error) {

	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	if gm.event.Status == events.Failure {
		msg.Text = fmt.Sprintf("Failed to roll back %s of %s",
			gm.event.RollbackRequest.App,
			gm.event.RollbackRequest.Env)

		msg.Embed.Description += fmt.Sprintf(":exclamation: *Error* :exclamation: \n%s\n", gm.event.StatusDesc)
		msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.event.RollbackRequest.Env))
		msg.Embed.Description += fmt.Sprintf(":clipboard: %s\n", gm.event.RollbackRequest.TargetSHA)

		msg.Embed.Color = 15158332

	} else {
		msg.Text = fmt.Sprintf(":arrow_backward: %s is rolling back %s on %s", gm.event.RollbackRequest.TriggeredBy, gm.event.RollbackRequest.App, gm.event.RollbackRequest.Env)

		msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.event.RollbackRequest.Env))
		msg.Embed.Description += fmt.Sprintf(":clipboard: %s\n", gm.event.RollbackRequest.TargetSHA)

		for _, gitopsRef := range gm.event.GitopsRefs {
			msg.Embed.Description += fmt.Sprintf(":paperclip: %s\n", discordCommitLink(gm.event.GitopsRepo, gitopsRef))
		}

		msg.Embed.Color = 3066993
	}

	return msg, nil
}

func MessageFromRollbackEvent(event *events.RollbackEvent) Message {
	return &gitopsRollbackMessage{
		event: event,
	}
}

func (gm *gitopsRollbackMessage) RepositoryName() string {
	return ""
}

func (gm *gitopsRollbackMessage) SHA() string {
	return ""
}
