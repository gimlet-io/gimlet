package notifications

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type AlertMessage struct {
	Alert         api.Alert
	ImChannelId   string
	DeploymentUrl string
}

func (am *AlertMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:        "",
		Attachments: []Attachment{},
	}

	msg.Text = fmt.Sprintf("ALERT: *%s %s* for pod *%s*", strings.Title(am.Alert.Type), strings.ToUpper(am.Alert.Status), am.Alert.ObjectName)
	if am.Alert.Status == model.RESOLVED {
		msg.Attachments = append(msg.Attachments,
			Attachment{
				Color: "#36a64f",
				Blocks: []Block{
					{
						Type: section,
						Text: &Text{
							Type: markdown,
							Text: fmt.Sprintf(":white_check_mark: %s alert resolved", am.Alert.Type),
						},
					},
				},
			})
	} else {
		msg.Attachments = append(msg.Attachments,
			Attachment{
				Color: "#FF0000",
				Blocks: []Block{
					{
						Type: section,
						Text: &Text{
							Type: markdown,
							Text: am.Alert.Text,
						},
						Accessory: &Accessory{
							Type: button,
							Text: &Text{
								Type: "plain_text",
								Text: "View",
							},
							Url:   am.DeploymentUrl,
							Style: "primary",
						},
					},
					{
						Type: section,
						Text: &Text{
							Type: markdown,
							Text: " ",
						},
						Accessory: &Accessory{
							Type: button,
							Text: &Text{
								Type: "plain_text",
								Text: "Silence",
							},
							Url:   am.DeploymentUrl,
							Style: "danger",
						},
					},
				},
			})
	}

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

func (am *AlertMessage) CustomChannel() string {
	return am.ImChannelId
}
