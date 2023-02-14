package notifications

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type fluxMessage struct {
	gitopsCommit *model.GitopsCommit
	gitopsRepo   string
	env          string
}

func (fm *fluxMessage) AsSlackMessage() (*slackMessage, error) {
	msg := &slackMessage{
		Text:   "",
		Blocks: []Block{},
	}

	switch fm.gitopsCommit.Status {
	case model.Progressing:
		if strings.Contains(fm.gitopsCommit.StatusDesc, "Health check passed") {
			msg.Text = fmt.Sprintf(":heavy_check_mark: Applied resources from %s are up and healthy", commitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		} else {
			msg.Text = fmt.Sprintf(":hourglass_flowing_sand: Applying gitops changes from %s", commitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		}
	case model.ValidationFailed:
		fallthrough
	case model.ReconciliationFailed:
		msg.Text = fmt.Sprintf(":exclamation: Gitops changes from %s failed to apply", commitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
	case model.HealthCheckFailed:
		msg.Text = fmt.Sprintf(":ambulance: Gitops changes from %s have health issues", commitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
	default:
		msg.Text = fmt.Sprintf("%s: %s", fm.gitopsCommit.Status, commitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
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

	var contextText string
	switch fm.gitopsCommit.Status {
	case model.ValidationFailed:
		fallthrough
	case model.ReconciliationFailed:
		fallthrough
	case model.HealthCheckFailed:
		contextText = fm.gitopsCommit.StatusDesc
	case model.Progressing:
		if strings.Contains(fm.gitopsCommit.StatusDesc, "Health check passed") {
			contextText = fm.gitopsCommit.StatusDesc
		}
	}

	if contextText != "" {
		msg.Blocks = append(msg.Blocks,
			Block{
				Type: contextString,
				Elements: []Text{
					{
						Type: markdown,
						Text: contextText,
					},
				},
			},
		)
	}

	return msg, nil
}

func (fm *fluxMessage) Env() string {
	return fm.env
}

func (fm *fluxMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (fm *fluxMessage) AsDiscordMessage() (*discordMessage, error) {

	msg := &discordMessage{
		Text: "Health check",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       3066993,
		},
	}

	switch fm.gitopsCommit.Status {
	case model.Progressing:
		if strings.Contains(fm.gitopsCommit.StatusDesc, "Health check passed") {
			msg.Embed.Description = fmt.Sprintf(":heavy_check_mark: Applied resources from %s are up and healthy", discordCommitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		} else {
			msg.Embed.Description = fmt.Sprintf(":hourglass_flowing_sand: Applying gitops changes from %s", discordCommitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		}
	case model.ValidationFailed:
		fallthrough
	case model.ReconciliationFailed:
		msg.Embed.Description = fmt.Sprintf(":exclamation: Gitops changes from %s failed to apply", discordCommitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		msg.Embed.Color = 15158332
	case model.HealthCheckFailed:
		msg.Embed.Description = fmt.Sprintf(":ambulance: Gitops changes from %s have health issues", discordCommitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
		msg.Embed.Color = 15158332
	default:
		msg.Embed.Description = fmt.Sprintf("%s: %s", fm.gitopsCommit.Status, discordCommitLink(fm.gitopsRepo, fm.gitopsCommit.Sha))
	}

	return msg, nil
}

func NewMessage(gitopsRepo string, gitopsCommit *model.GitopsCommit, env string) Message {
	return &fluxMessage{
		gitopsCommit: gitopsCommit,
		gitopsRepo:   gitopsRepo,
		env:          env,
	}
}

func (fm *fluxMessage) RepositoryName() string {
	return ""
}

func (fm *fluxMessage) SHA() string {
	return ""
}
