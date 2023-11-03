package notifications

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const githubCommitLinkFormatForDiscord = "[%s](https://github.com/%s/commit/%s)"

type DiscordProvider struct {
	Token          string
	ChannelID      string
	ChannelMapping map[string]string
}

type discordMessage struct {
	Text  string                  `json:"text"`
	Embed *discordgo.MessageEmbed `json:"embed"`
}

func (s *DiscordProvider) send(msg Message) error {
	discordBot, err := discordgo.New("Bot " + s.Token)
	if err != nil {
		return fmt.Errorf("error creating Discord session, %s", err)
	}

	discordMessage, err := msg.AsDiscordMessage()
	if err != nil {
		return fmt.Errorf("cannot create slack message: %s", err)
	}

	channel := s.ChannelID
	if ch, ok := s.ChannelMapping[msg.Env()]; ok {
		channel = ch
	}
	s.ChannelID = channel

	return s.post(discordBot, discordMessage)
}

func (s *DiscordProvider) post(d *discordgo.Session, msg *discordMessage) error {
	_, err := d.ChannelMessageSend(s.ChannelID, msg.Text)
	if err != nil {
		return err
	}

	d.ChannelMessageSendEmbed(s.ChannelID, msg.Embed)
	if err != nil {
		return err
	}

	return nil
}

func discordCommitLink(repo string, ref string) string {
	if len(ref) < 8 {
		return ""
	}
	return fmt.Sprintf(githubCommitLinkFormatForDiscord, ref[0:7], repo, ref)
}
