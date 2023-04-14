package alert

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
)

var (
	encryptionKey    = "the-key-has-to-be-32-bytes-long!"
	encryptionKeyNew = ""
)

func TestTrackPods_UpdatesPodState(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pods := []*api.Pod{&pod1}

	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
	p.TrackPods(pods)

	expectedPods := []model.Pod{
		{Name: "ns1/pod1", Status: "Running"},
	}
	for _, pod := range expectedPods {
		p, _ := store.Pod(pod.Name)

		assert.Equal(t, p.Name, pod.Name)
		assert.Equal(t, p.Status, pod.Status)
	}
}

func TestTrackPods_ErrorStateCreatesPendingAlert(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pod3 := api.Pod{Namespace: "ns2", Name: "pod3", Status: "Pending"}
	pods := []*api.Pod{&pod1, &pod2, &pod3}

	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
	p.TrackPods(pods)

	expectedAlerts := []model.Alert{
		{ObjectName: "ns1/pod2", ObjectType: "pod", Status: "Pending"},
	}
	for _, alert := range expectedAlerts {
		a, _ := store.Alerts(alert.ObjectName, alert.ObjectType)

		assert.Equal(t, a[0].Status, alert.Status)
	}
}

func TestTrackEvents(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	event1 := api.Event{Namespace: "ns1", Name: "pod1"}
	event2 := api.Event{Namespace: "ns1", Name: "pod2"}
	events := []api.Event{event1, event2}

	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
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
		{ObjectName: "ns1/pod1", ObjectType: "event", Status: "Pending"},
		{ObjectName: "ns1/pod2", ObjectType: "event", Status: "Pending"},
	}
	for _, alert := range expectedAlerts {
		a, _ := store.Alerts(alert.ObjectName, alert.ObjectType)

		assert.Equal(t, alert.Status, a[0].Status)
	}
}

func TestEvaluatePendingAlerts(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	currentTime := time.Now()
	dummyNotificationsManager := notifications.NewDummyManager()
	a := NewAlertStateManager(dummyNotificationsManager, *store, 2)
	alert1 := model.Alert{ObjectName: "n/p1", ObjectType: "pod", Status: "Pending", LastStateChange: currentTime.Add(-1 * time.Minute).Unix()}
	alert2 := model.Alert{ObjectName: "n/p2", ObjectType: "pod", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
	alert3 := model.Alert{ObjectName: "n/e1", ObjectType: "event", Status: "Pending", LastStateChange: currentTime.Add(-4 * time.Minute).Unix()}
	alert4 := model.Alert{ObjectName: "n/e2", ObjectType: "event", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
	alerts := []model.Alert{alert1, alert2, alert3, alert4}
	for _, alert := range alerts {
		store.CreateAlert(&alert)
	}

	pod1 := model.Pod{Name: "n/p1", Status: "ImagePullBackOff", StatusDesc: "blabla"}
	store.SaveOrUpdatePod(&pod1)
	pod2 := model.Pod{Name: "n/p2", Status: "ImagePullBackOff", StatusDesc: "blabla"}
	store.SaveOrUpdatePod(&pod2)
	event1 := model.KubeEvent{Name: "n/e1", Status: "Failed", StatusDesc: "blabla", Count: 5}
	store.SaveOrUpdateKubeEvent(&event1)
	event2 := model.KubeEvent{Name: "n/e2", Status: "Failed", StatusDesc: "blabla", Count: 7}
	store.SaveOrUpdateKubeEvent(&event2)

	a.evaluatePendingAlerts()

	expected := []model.Alert{
		{ObjectName: "n/p1", ObjectType: "pod", Status: "Pending"},
		{ObjectName: "n/p2", ObjectType: "pod", Status: "Firing"},
		{ObjectName: "n/e1", ObjectType: "event", Status: "Pending"},
		{ObjectName: "n/e2", ObjectType: "event", Status: "Firing"},
	}
	for _, alert := range expected {
		a, _ := store.Alerts(alert.ObjectName, alert.ObjectType)
		assert.Equal(t, alert.Status, a[0].Status)
	}
}

func TestPodFailedMessage(t *testing.T) {
	msgPodFailed := notifications.AlertMessage{
		Alert: model.Alert{
			ObjectType: "pod",
			ObjectName: "ns1/pod1",
			// StatusDesc: "Container failed",
		},
	}

	discordMsg, err := msgPodFailed.AsDiscordMessage()
	assert.Nil(t, err)

	assert.Contains(t, "pod ns1/pod1 failed", discordMsg.Text)

	slackMsg, err := msgPodFailed.AsSlackMessage()
	assert.Nil(t, err)

	assert.Contains(t, "pod ns1/pod1 failed", slackMsg.Text)
}

// func TestNotificationSending(t *testing.T) {
// 	t.Skip("Skipping notification sending")
// 	store := store.NewTest(encryptionKey, encryptionKeyNew)
// 	defer func() {
// 		store.Close()
// 	}()

// 	notificationsManager := notifications.NewManager()

// 	notificationsManager.AddProvider(&notifications.DiscordProvider{
// 		Token:     "",
// 		ChannelID: "",
// 	})

// 	go notificationsManager.Run()

// 	p := NewAlertStateManager(notificationsManager, *store, 2)

// 	var thresholds []threshold
// 	currentTime := time.Now()
// 	alert1 := model.Alert{Name: "n/p2", Type: "pod", Status: "Pending", StatusDesc: "Back-off pulling image", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
// 	thresholds = append(thresholds, ToThreshold(&alert1, p.waitTime, 6, 1))

// 	p.setFiringState(thresholds)

// 	time.Sleep(5 * time.Second)
// }
