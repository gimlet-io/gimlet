package alert

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
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

	// dummyNotificationsManager := notifications.NewDummyManager()
	// agentHub := streaming.NewAgentHub(&config.Config{})
	// pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	// pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	// pod3 := api.Pod{Namespace: "ns2", Name: "pod3", Status: "Pending"}
	// pods := []*api.Pod{&pod1, &pod2, &pod3}

	// p := NewAlertStateManager(dummyNotificationsManager, agentHub, *store, 2)
	// p.TrackPods(pods)

	// expected := []model.Pod{
	// 	{Name: "ns1/pod1"},
	// 	{Name: "ns1/pod2"},
	// 	{Name: "ns2/pod3"},
	// }
	// for _, pod := range expected {
	// 	p, _ := store.Alert(pod.Name, "pod")

	// 	assert.Equal(t, p.Name, pod.Name)
	// }
}

func TestTrackEvents(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	// dummyNotificationsManager := notifications.NewDummyManager()
	// agentHub := streaming.NewAgentHub(&config.Config{})
	// event1 := api.Event{Namespace: "ns1", Name: "pod1"}
	// event2 := api.Event{Namespace: "ns1", Name: "pod2"}
	// events := []api.Event{event1, event2}

	// p := NewAlertStateManager(dummyNotificationsManager, agentHub, *store, 2)
	// p.TrackEvents(events)

	// expected := []model.KubeEvent{
	// 	{Name: "ns1/pod1"},
	// 	{Name: "ns1/pod2"},
	// }
	// for _, event := range expected {
	// 	e, _ := store.Alert(event.Name, "event")

	// 	assert.Equal(t, event.Name, e.Name)
	// }
}

// TODO write
func TestSetFiringState(t *testing.T) {

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

// TODO rewrite
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

	pod1 := model.Pod{Name: "ns1/pod1", Status: "Error", StatusDesc: "Back-off pulling image"}
	pod2 := model.Pod{Name: "ns1/pod2", Status: "ErrImagePull", StatusDesc: "Back-off pulling image"}

	store.SaveOrUpdatePod(&pod1)
	store.SaveOrUpdatePod(&pod2)

	p.setFiringState([]threshold{})

	time.Sleep(5 * time.Second)
}
