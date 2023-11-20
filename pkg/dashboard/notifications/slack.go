package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

const markdown = "mrkdwn"
const section = "section"
const contextString = "context"
const button = "button"

const githubCommitLinkFormat = "<https://github.com/%s/commit/%s|%s>"
const bitbucketServerLinkFormat = "<http://%s/projects/%s/repos/%s/commits/%s|%s>"

type SlackProvider struct {
	Token          string
	DefaultChannel string
	ChannelMapping map[string]string
}

type slackMessage struct {
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Block struct {
	Type      string     `json:"type"`
	Text      *Text      `json:"text,omitempty"`
	Accessory *Accessory `json:"accessory,omitempty"`
	Elements  []Text     `json:"elements,omitempty"`
}

type Attachment struct {
	Color  string  `json:"color"`
	Blocks []Block `json:"blocks,omitempty"`
}

type Accessory struct {
	Text *Text  `json:"text"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *SlackProvider) send(msg Message) error {
	slackMessage, err := msg.AsSlackMessage()
	if err != nil {
		return fmt.Errorf("cannot create slack message: %s", err)
	}

	if slackMessage == nil {
		return nil
	}

	slackMessage.Channel = s.channel(msg)

	return s.post(slackMessage)
}

func (s *SlackProvider) channel(msg Message) string {
	if msg.CustomChannel() != "" {
		return msg.CustomChannel()
	}

	if ch, ok := s.ChannelMapping[msg.Env()]; ok {
		return ch
	}

	return s.DefaultChannel
}

func (s *SlackProvider) post(msg *slackMessage) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(msg)
	if err != nil {
		logrus.Printf("Could encode message to slack: %v", err)
		return err
	}

	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", b)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	req = req.WithContext(context.TODO())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logrus.Printf("could not post to slack: %v", err)
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	var parsed map[string]interface{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return fmt.Errorf("cannot parse slack response: %s", err)
	}
	if val, ok := parsed["ok"]; ok {
		if val != true {
			logrus.Infof("Slack response: %s", string(body))
		}
	} else {
		logrus.Infof("Slack response: %s", string(body))
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("could not post to slack, status: %d", res.StatusCode)
	}

	return nil
}

func commitLink(repo string, ref string) string {
	if len(ref) < 8 {
		return ""
	}
	return fmt.Sprintf(githubCommitLinkFormat, repo, ref, ref[0:7])
}
