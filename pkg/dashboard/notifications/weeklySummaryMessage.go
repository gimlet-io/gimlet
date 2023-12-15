package notifications

import (
	"fmt"
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
)

type weeklySummaryOpts struct {
	deploys                int
	rollbacks              int
	mostTriggeredBy        string
	alertSeconds           int
	alertsPercentageChange float64
	serviceLag             map[string]float64
	repos                  []string
	scmUrl                 string
}

type weeklySummaryMessage struct {
	opts weeklySummaryOpts
}

func (ws *weeklySummaryMessage) AsSlackMessage() (*slackMessage, error) {
	t := time.Now()

	msg := &slackMessage{
		Blocks: []Block{
			{
				Type: header,
				Text: &Text{
					Type: "plain_text",
					Text: ":chart_with_upwards_trend:  Gimlet weekly summary  :chart_with_upwards_trend:",
				},
			},
			{
				Type: contextString,
				Elements: []Text{
					{
						Type: markdown,
						Text: fmt.Sprintf("*%s %d, %d*  |  Gimlet Team Announcements", t.Month().String(), t.Day(), t.Year()),
					},
				},
			},
			{
				Type: divider,
			},
			{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: ":clipboard: *DEPLOYMENTS* :clipboard:",
				},
			},
			{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: fmt.Sprintf("There were *%d* application rollouts, and *%d* rollbacks total.", ws.opts.deploys, ws.opts.rollbacks),
				},
			},
			{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: fmt.Sprintf(":trophy: *@%s* did the most deploys.", ws.opts.mostTriggeredBy),
				},
			},
			{
				Type: divider,
			},
			{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: ":rotating_light: *TOTAL TIME OF ALERTS* :rotating_light:",
				},
			},
			{
				Type: section,
				Text: &Text{
					Type: markdown,
					Text: fmt.Sprintf("There were *%d* seconds of alerts total.", ws.opts.alertSeconds),
				},
			},
		},
	}

	if !math.IsNaN(ws.opts.alertsPercentageChange) && !math.IsInf(ws.opts.alertsPercentageChange, 1) {
		change := "more"
		if math.Signbit(ws.opts.alertsPercentageChange) {
			change = "less"
		}

		msg.Blocks = append(msg.Blocks, Block{
			Type: section,
			Text: &Text{
				Type: markdown,
				Text: fmt.Sprintf("This is *%.2f%%* %s than the previous week.", ws.opts.alertsPercentageChange, change),
			},
		})
	}

	msg.Blocks = append(msg.Blocks, Block{
		Type: divider,
	})

	msg.Blocks = append(msg.Blocks, lag(ws.opts.serviceLag)...)
	msg.Blocks = append(msg.Blocks, Block{
		Type: divider,
	})

	msg.Blocks = append(msg.Blocks, repos(ws.opts.repos, ws.opts.scmUrl)...)
	msg.Blocks = append(msg.Blocks, Block{
		Type: divider,
	})

	msg.Blocks = append(msg.Blocks, Block{
		Type: contextString,
		Elements: []Text{
			{
				Type: markdown,
				Text: ":cocktail: This newsletter was sent by *<https://gimlet.io|Gimlet>*.",
			},
		},
	})

	return msg, nil
}

func lag(lagSeconds map[string]float64) (b []Block) {
	b = append(b, Block{
		Type: section,
		Text: &Text{
			Type: markdown,
			Text: ":hourglass_flowing_sand: *MEAN LAG* :hourglass_flowing_sand:",
		},
	})

	if len(lagSeconds) == 0 {
		b = append(b, Block{
			Type: section,
			Text: &Text{
				Type: markdown,
				Text: "No mean lag between staging and production.",
			},
		})

		return
	}

	for app, seconds := range lagSeconds {
		if seconds == 0 {
			continue
		}

		b = append(b, Block{
			Type: section,
			Text: &Text{
				Type: markdown,
				Text: fmt.Sprintf("_Production_ is lagging behind _staging_ with *%v* seconds on *%s* app.", math.Abs(seconds), app),
			},
		})
	}

	return
}

func repos(repos []string, scmUrl string) (b []Block) {
	b = append(b, Block{
		Type: section,
		Text: &Text{
			Type: markdown,
			Text: ":package: *REPOSITORIES WHERE STAGING IS BEHIND PRODUCTION* :package:",
		},
	})

	if len(repos) == 0 {
		b = append(b, Block{
			Type: section,
			Text: &Text{
				Type: markdown,
				Text: "There are no repos where staging is behind production.",
			},
		})
	}

	for _, repo := range repos {
		b = append(b, Block{
			Type: section,
			Text: &Text{
				Type: markdown,
				Text: fmt.Sprintf("*<%s/%s|%s>*", scmUrl, repo, repo),
			},
		})
	}

	return
}

func WeeklySummary(
	deploys, rollbacks int,
	mostTriggeredBy string,
	alertSeconds int,
	alertsPercentageChange float64,
	serviceLag map[string]float64,
	repos []string,
	scmUrl string,
) Message {
	return &weeklySummaryMessage{
		opts: weeklySummaryOpts{
			deploys:                deploys,
			rollbacks:              rollbacks,
			mostTriggeredBy:        mostTriggeredBy,
			alertSeconds:           alertSeconds,
			alertsPercentageChange: alertsPercentageChange,
			serviceLag:             serviceLag,
			repos:                  repos,
			scmUrl:                 scmUrl,
		},
	}
}

func (ws *weeklySummaryMessage) Env() string {
	return ""
}

func (ws *weeklySummaryMessage) AsStatus() (*status, error) {
	return nil, nil
}

func (ws *weeklySummaryMessage) AsDiscordMessage() (*discordMessage, error) {
	msg := &discordMessage{
		Text: "",
		Embed: &discordgo.MessageEmbed{
			Type:        "article",
			Description: "",
			Color:       0,
		},
	}

	msg.Text = "TODO"
	msg.Embed.Color = 15158332

	return msg, nil
}

func (ws *weeklySummaryMessage) RepositoryName() string {
	return ""
}

func (ws *weeklySummaryMessage) SHA() string {
	return ""
}

func (ws *weeklySummaryMessage) CustomChannel() string {
	return ""
}
