package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type AlertMessage struct {
	Alert api.Alert
}

func (am *AlertMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("%s %s %s", am.Alert.ObjectName, am.Alert.Type, am.Alert.Status)
	desc := fmt.Sprintf(":exclamation: %s", am.Alert.Text)
	if am.Alert.Status == model.RESOLVED {
		desc = ":white_check_mark: Running"
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
				{
					Type: markdown,
					Text: desc,
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

	desc := fmt.Sprintf(":exclamation: %s", am.Alert.Text)
	color := 15158332
	if am.Alert.Status == model.RESOLVED {
		desc = ":white_check_mark: Running"
		color = 3066993
	}

	msg.Text = fmt.Sprintf("%s %s %s", am.Alert.ObjectName, am.Alert.Type, am.Alert.Status)
	msg.Embed.Description += desc
	msg.Embed.Color = color

	return msg, nil
}

func (am *AlertMessage) RepositoryName() string {
	return ""
}

func (am *AlertMessage) SHA() string {
	return ""
}
