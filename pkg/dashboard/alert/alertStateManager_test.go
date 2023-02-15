package alert

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
)

func TestTrackPods(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	agentHub := streaming.NewAgentHub(&config.Config{})
	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pod3 := api.Pod{Namespace: "ns2", Name: "pod3", Status: "Pending"}
	pods := []*api.Pod{&pod1, &pod2, &pod3}

	p := NewAlertStateManager(dummyNotificationsManager, agentHub, *store, 2)
	p.TrackPods(pods)

	expectedPods := []model.Pod{
		{Name: "ns1/pod1"},
		{Name: "ns1/pod2"},
		{Name: "ns2/pod3"},
	}
	for _, pod := range expectedPods {
		p, _ := store.Pod(pod.Name)

		assert.Equal(t, p.Name, pod.Name)
	}

	expectedAlerts := []model.Alert{
		{Name: "ns1/pod2", Type: "pod", Status: "Pending"},
	}
	for _, alert := range expectedAlerts {
		a, _ := store.Alert(alert.Name, alert.Type)

		assert.Equal(t, a.Status, alert.Status)
	}
}

func TestTrackEvents(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	agentHub := streaming.NewAgentHub(&config.Config{})
	event1 := api.Event{Namespace: "ns1", Name: "pod1"}
	event2 := api.Event{Namespace: "ns1", Name: "pod2"}
	events := []api.Event{event1, event2}

	p := NewAlertStateManager(dummyNotificationsManager, agentHub, *store, 2)
	p.TrackEvents(events)

	expectedEvents := []model.KubeEvent{
		{Name: "ns1/pod1"},
		{Name: "ns1/pod2"},
	}
	for _, event := range expectedEvents {
		e, _ := store.KubeEvent(event.Name)

		assert.Equal(t, event.Name, e.Name)
	}

	expectedAlerts := []model.Alert{
		{Name: "ns1/pod1", Type: "event", Status: "Pending"},
		{Name: "ns1/pod2", Type: "event", Status: "Pending"},
	}
	for _, alert := range expectedAlerts {
		a, _ := store.Alert(alert.Name, alert.Type)

		assert.Equal(t, alert.Status, a.Status)
	}
}

func TestSetFiringState(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	currentTime := time.Now()
	dummyNotificationsManager := notifications.NewDummyManager()
	agentHub := streaming.NewAgentHub(&config.Config{})
	a := NewAlertStateManager(dummyNotificationsManager, agentHub, *store, 2)
	alert1 := model.Alert{Name: "n/p1", Type: "pod", Status: "Pending", LastStateChange: currentTime.Add(-1 * time.Minute).Unix()}
	alert2 := model.Alert{Name: "n/p2", Type: "pod", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
	alert3 := model.Alert{Name: "n/p1", Type: "event", Status: "Pending", LastStateChange: currentTime.Add(-4 * time.Minute).Unix(), Count: 5}
	alert4 := model.Alert{Name: "n/p2", Type: "event", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix(), Count: 7}
	alerts := []model.Alert{alert1, alert2, alert3, alert4}

	for _, alert := range alerts {
		store.SaveOrUpdateAlert(&alert)
	}

	var thresholds []threshold
	for _, alert := range alerts {
		thresholds = append(thresholds, ToThreshold(&alert, a.waitTime, 6, 1))
	}

	err := a.setFiringState(thresholds)
	assert.Nil(t, err)

	expected := []model.Alert{
		{Name: "n/p1", Type: "pod", Status: "Pending"},
		{Name: "n/p2", Type: "pod", Status: "Firing"},
		{Name: "n/p1", Type: "event", Status: "Pending"},
		{Name: "n/p2", Type: "event", Status: "Firing"},
	}
	for _, alert := range expected {
		a, _ := store.Alert(alert.Name, alert.Type)
		assert.Equal(t, alert.Status, a.Status)
	}
}

func TestPodFailedMessage(t *testing.T) {
	msgPodFailed := notifications.AlertMessage{
		Alert: model.Alert{
			Type:       "pod",
			Name:       "ns1/pod1",
			StatusDesc: "Container failed",
		},
	}

	discordMsg, err := msgPodFailed.AsDiscordMessage()
	assert.Nil(t, err)

	assert.Contains(t, "pod ns1/pod1 failed", discordMsg.Text)

	slackMsg, err := msgPodFailed.AsSlackMessage()
	assert.Nil(t, err)

	assert.Contains(t, "pod ns1/pod1 failed", slackMsg.Text)
}

func TestNotificationSending(t *testing.T) {
	t.Skip("Skipping notification sending")
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	notificationsManager := notifications.NewManager()
	agentHub := streaming.NewAgentHub(&config.Config{})

	notificationsManager.AddProvider(&notifications.DiscordProvider{
		Token:     "",
		ChannelID: "",
	})

	go notificationsManager.Run()

	p := NewAlertStateManager(notificationsManager, agentHub, *store, 2)

	var thresholds []threshold
	currentTime := time.Now()
	alert1 := model.Alert{Name: "n/p2", Type: "pod", Status: "Pending", StatusDesc: "Back-off pulling image", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
	thresholds = append(thresholds, ToThreshold(&alert1, p.waitTime, 6, 1))

	p.setFiringState(thresholds)

	time.Sleep(5 * time.Second)
}
