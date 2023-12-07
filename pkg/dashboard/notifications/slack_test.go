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
			lagSeconds: map[string]int64{
				"getting-started-app": 2600,
				"remix-test-app ":     110,
			},
			repos: []string{"gimlet-io/expressjs-test-app", "gimlet-io/reactjs-test-app"},
		},
	}
	manager.Broadcast(wsMessage)
	time.Sleep(5 * time.Second)
}
