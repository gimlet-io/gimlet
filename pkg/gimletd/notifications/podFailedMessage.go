package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type podFailedMessage struct {
	pod model.Pod
}

func (pm *podFailedMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("Pod %s failed with status %s", pm.pod.Name, pm.pod.Status)
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
					Text: fmt.Sprintf(":exclamation: %s", pm.pod.StatusDesc),
				},
			},
		},
	)

	return msg, nil
}

func (pm *podFailedMessage) Env() string {
	return ""
}

func (pm *podFailedMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (pm *podFailedMessage) AsDiscordMessage() (*discordMessage, error) {
	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	msg.Text = fmt.Sprintf("Pod %s failed with status %s", pm.pod.Name, pm.pod.Status)
	msg.Embed.Description += fmt.Sprintf(":exclamation: %s", pm.pod.StatusDesc)
	msg.Embed.Color = 15158332

	return msg, nil
}

func MessageFromFailedPod(pod model.Pod) Message {
	return &podFailedMessage{
		pod: pod,
	}
}

func (pm *podFailedMessage) RepositoryName() string {
	return ""
}

func (pm *podFailedMessage) SHA() string {
	return ""
}
