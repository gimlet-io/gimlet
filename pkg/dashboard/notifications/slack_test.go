package notifications

import (
	"testing"
	"time"
)

func Test_Slack(t *testing.T) {
	manager := NewManager()
	slack := &SlackProvider{
		Token:          "",
		DefaultChannel: "",
	}

	manager.AddProvider(slack)
	go manager.Run()
	wsMessage := &weeklySummaryMessage{
		opts: weeklySummaryOpts{
			deploys:         45,
			rollbacks:       8,
			mostTriggeredBy: "policy",
			alertSeconds:    370,
			alertChange:     -17,
			lagSeconds:      4300,
			repos:           []string{"gimlet/getting-started-app", "gimlet/onechart"},
		},
	}
	manager.Broadcast(wsMessage)
	time.Sleep(5 * time.Second)
}
