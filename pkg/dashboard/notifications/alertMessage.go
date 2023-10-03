package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type AlertMessage struct {
	Alert model.Alert
}

func (am *AlertMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("%s %s failed", am.Alert.Type, am.Alert.ObjectName)
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
					Text: fmt.Sprintf(":exclamation: %s", "TODO: related object and alert type"),
				},
			},
		},
	)

	return msg, nil
}

func (am *AlertMessage) Env() string {
	return ""
}

func (am *AlertMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (am *AlertMessage) AsDiscordMessage() (*discordMessage, error) {
	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	msg.Text = fmt.Sprintf("%s %s failed", am.Alert.Type, am.Alert.ObjectName)
	msg.Embed.Description += fmt.Sprintf(":exclamation: %s", "TODO: related object and alert type")
	msg.Embed.Color = 15158332

	return msg, nil
}

func (am *AlertMessage) RepositoryName() string {
	return ""
}

func (am *AlertMessage) SHA() string {
	return ""
}
