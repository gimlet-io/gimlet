package notifications

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
)

type gitopsDeleteMessage struct {
	result model.Result
}

func (gm *gitopsDeleteMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	if gm.result.Status == model.Failure {
		msg.Text = fmt.Sprintf("Failed to delete %s of %s", gm.result.Manifest.App, gm.result.Manifest.Env)
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
						Text: fmt.Sprintf(":exclamation: *Error* :exclamation: \n%s", gm.result.StatusDesc),
					},
				},
			},
		)
	} else {
		msg.Text = fmt.Sprintf("Policy based deletion of %s on %s", gm.result.Manifest.App, gm.result.Manifest.Env)
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
					{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.result.Manifest.Env))},
					{Type: markdown, Text: fmt.Sprintf(":paperclip: %s", commitLink(gm.result.GitopsRepo, gm.result.GitopsRef))},
				},
			},
		)
	}

	return msg, nil
}

func (gm *gitopsDeleteMessage) Env() string {
	return gm.result.Manifest.Env
}

func (gm *gitopsDeleteMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (gm *gitopsDeleteMessage) AsDiscordMessage() (*discordMessage, error) {

	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	if gm.result.Status == model.Failure {
		msg.Text = fmt.Sprintf("Failed to delete %s of %s", gm.result.Manifest.App, gm.result.Manifest.Env)
		msg.Embed.Description += fmt.Sprintf(":exclamation: *Error* :exclamation: \n%s", gm.result.StatusDesc)
		msg.Embed.Color = 15158332
	} else {
		msg.Text = fmt.Sprintf("Policy based deletion of %s on %s", gm.result.Manifest.App, gm.result.Manifest.Env)
		msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.result.Manifest.Env))
		msg.Embed.Description += fmt.Sprintf(":paperclip: %s\n", discordCommitLink(gm.result.GitopsRepo, gm.result.GitopsRef))

		msg.Embed.Color = 3066993
	}

	return msg, nil
}

func MessageFromDeleteEvent(result model.Result) Message {
	return &gitopsDeleteMessage{
		result: result,
	}
}

func (gm *gitopsDeleteMessage) RepositoryName() string {
	return ""
}

func (gm *gitopsDeleteMessage) SHA() string {
	return ""
}

func (gm *gitopsDeleteMessage) CustomChannel() string {
	return ""
}
