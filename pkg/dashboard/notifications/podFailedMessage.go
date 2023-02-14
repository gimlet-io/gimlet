package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type PodFailedMessage struct {
	Pod model.Pod
}

func (pm *PodFailedMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("Pod %s failed with status %s", pm.Pod.Name, pm.Pod.Status)
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
					Text: fmt.Sprintf(":exclamation: %s", pm.Pod.StatusDesc),
				},
			},
		},
	)

	return msg, nil
}

func (pm *PodFailedMessage) Env() string {
	return ""
}

func (pm *PodFailedMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (pm *PodFailedMessage) AsDiscordMessage() (*discordMessage, error) {
	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	msg.Text = fmt.Sprintf("Pod %s failed with status %s", pm.Pod.Name, pm.Pod.Status)
	msg.Embed.Description += fmt.Sprintf(":exclamation: %s", pm.Pod.StatusDesc)
	msg.Embed.Color = 15158332

	return msg, nil
}

func MessageFromFailedPod(pod model.Pod) Message {
	return &PodFailedMessage{
		Pod: pod,
	}
}

func (pm *PodFailedMessage) RepositoryName() string {
	return ""
}

func (pm *PodFailedMessage) SHA() string {
	return ""
}
