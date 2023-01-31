package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
)

type WarningEventMessage struct {
	Event api.Event
}

func (we *WarningEventMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("%s %s", we.Event.Name, we.Event.Status)
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
					Text: fmt.Sprintf(":exclamation: %s", we.Event.StatusDesc),
				},
			},
		},
	)

	return msg, nil
}

func (we *WarningEventMessage) Env() string {
	return ""
}

func (we *WarningEventMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (we *WarningEventMessage) AsDiscordMessage() (*discordMessage, error) {
	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	msg.Text = fmt.Sprintf("%s %s", we.Event.Name, we.Event.Status)
	msg.Embed.Description += fmt.Sprintf(":exclamation: %s", we.Event.StatusDesc)
	msg.Embed.Color = 15158332

	return msg, nil
}

func MessageFromWarningEvent(event api.Event) Message {
	return &WarningEventMessage{
		Event: event,
	}
}

func (we *WarningEventMessage) RepositoryName() string {
	return ""
}

func (we *WarningEventMessage) SHA() string {
	return ""
}
