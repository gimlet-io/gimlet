package notifications

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dx"
)

type gitopsRollbackMessage struct {
	event           model.Event
	rollbackRequest dx.RollbackRequest
}

func (gm *gitopsRollbackMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	msg.Text = fmt.Sprintf("ROLLBACK: *%s* is rolling back *%s* on %s", gm.rollbackRequest.TriggeredBy, gm.rollbackRequest.App, gm.rollbackRequest.Env)
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
				{Type: markdown, Text: fmt.Sprintf(":dart: %s", strings.Title(gm.rollbackRequest.Env))},
				{Type: markdown, Text: fmt.Sprintf(":clipboard: %s", gm.rollbackRequest.TargetSHA)},
			},
		},
	)
	for _, result := range gm.event.Results {
		msg.Blocks[len(msg.Blocks)-1].Elements = append(
			msg.Blocks[len(msg.Blocks)-1].Elements,
			Text{Type: markdown, Text: fmt.Sprintf(":paperclip: %s", commitLink(result.GitopsRepo, result.GitopsRef))},
		)
	}
	if len(msg.Blocks[len(msg.Blocks)-1].Elements) > 10 {
		msg.Blocks[len(msg.Blocks)-1].Elements = msg.Blocks[len(msg.Blocks)-1].Elements[:10]
	}

	return msg, nil
}

func (gm *gitopsRollbackMessage) Env() string {
	return gm.rollbackRequest.Env
}

func (gm *gitopsRollbackMessage) AsStatus() (*status, error) {
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

	msg.Text = fmt.Sprintf(":arrow_backward: %s is rolling back %s on %s", gm.rollbackRequest.TriggeredBy, gm.rollbackRequest.App, gm.rollbackRequest.Env)

	msg.Embed.Description += fmt.Sprintf(":dart: %s\n", strings.Title(gm.rollbackRequest.Env))
	msg.Embed.Description += fmt.Sprintf(":clipboard: %s\n", gm.rollbackRequest.TargetSHA)

	for _, result := range gm.event.Results {
		msg.Embed.Description += fmt.Sprintf(":paperclip: %s\n", discordCommitLink(result.GitopsRepo, result.GitopsRef))
	}

	msg.Embed.Color = 3066993

	return msg, nil
}

func MessageFromRollbackEvent(event model.Event) (Message, error) {

	var rollbackRequest dx.RollbackRequest
	err := json.Unmarshal([]byte(event.Blob), &rollbackRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot parse rollback request with id: %s", event.ID)
	}

	return &gitopsRollbackMessage{
		event:           event,
		rollbackRequest: rollbackRequest,
	}, nil
}

func (gm *gitopsRollbackMessage) RepositoryName() string {
	return ""
}

func (gm *gitopsRollbackMessage) SHA() string {
	return ""
}

func (gm *gitopsRollbackMessage) CustomChannel() string {
	return ""
}
