package alert

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/stretchr/testify/assert"
)

var (
	encryptionKey    = "the-key-has-to-be-32-bytes-long!"
	encryptionKeyNew = ""
)

func TestTrackPod_imagePullBackOff(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"ImagePullBackOff": imagePullBackOffThreshold{
				waitTime: 0,
			}},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ImagePullBackOff",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.PENDING, relatedAlerts[0].Status)

	alertStateManager.evaluatePendingAlerts()
	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.FIRING, relatedAlerts[0].Status)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    model.POD_RUNNING,
	}, "", "")

	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.RESOLVED, relatedAlerts[0].Status)
}

func TestTrackPod_crashLoopBackOff(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"CrashLoopBackOff": crashLoopBackOffThreshold{
				waitTime: 0,
			}},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "CrashLoopBackOff",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.PENDING, relatedAlerts[0].Status)

	alertStateManager.evaluatePendingAlerts()
	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.FIRING, relatedAlerts[0].Status)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    model.POD_RUNNING,
	}, "", "")

	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.RESOLVED, relatedAlerts[0].Status)
}

func TestTrackPod_OOMilled(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	alertStateManager := NewAlertStateManager(
		notifications.NewDummyManager(),
		nil,
		*store,
		0,
		map[string]threshold{
			"OOMKilled":        oomKilledcrashLoopBackOffThreshold{},
			"CrashLoopBackOff": crashLoopBackOffThreshold{},
		},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "OOMKilled",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.PENDING, relatedAlerts[0].Status)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "CrashLoopBackOff",
	}, "", "")

	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, "oomKilledcrashLoopBackOffThreshold", relatedAlerts[0].Type)

	alertStateManager.evaluatePendingAlerts()
	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.FIRING, relatedAlerts[0].Status)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    model.POD_RUNNING,
	}, "", "")

	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.RESOLVED, relatedAlerts[0].Status)
}

func TestTrackPod_createContainerConfigError(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"CreateContainerConfigError": createContainerConfigErrorThreshold{
				waitTime: 0,
			}},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "CreateContainerConfigError",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.PENDING, relatedAlerts[0].Status)

	alertStateManager.evaluatePendingAlerts()
	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.FIRING, relatedAlerts[0].Status)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    model.POD_RUNNING,
	}, "", "")

	relatedAlerts, _ = store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.RESOLVED, relatedAlerts[0].Status)
}

func TestTrackPod_deleted(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"ImagePullBackOff": imagePullBackOffThreshold{
				waitTime: 0,
			}},
		"",
	)

	pod := &api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ImagePullBackOff",
	}

	alertStateManager.TrackPod(pod, "", "")

	alertStateManager.DeletePod(pod.FQN())

	relatedAlerts, _ := store.RelatedAlerts(pod.FQN())
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, model.RESOLVED, relatedAlerts[0].Status)
}

func TestTrackPod_from_pending_to_imagePullBackoff(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"Pending": pendingThreshold{
				waitTime: 0,
			},
			"ImagePullBackOff": imagePullBackOffThreshold{
				waitTime: 0,
			}},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "Pending",
	}, "", "")

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ImagePullBackOff",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 2, len(relatedAlerts))
	assert.Equal(t, "pendingThreshold", relatedAlerts[0].Type)
	assert.Equal(t, "Resolved", relatedAlerts[0].Status)
	assert.Equal(t, "imagePullBackOffThreshold", relatedAlerts[1].Type)
	assert.Equal(t, "Pending", relatedAlerts[1].Status)
}

func TestPodsTrack(t *testing.T) {
	store := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()

	alertStateManager := NewAlertStateManager(
		dummyNotificationsManager,
		nil,
		*store,
		0,
		map[string]threshold{
			"ImagePullBackOff": imagePullBackOffThreshold{
				waitTime: 0,
			}},
		"",
	)

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ImagePullBackOff",
	}, "", "")

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ErrImagePull",
	}, "", "")

	alertStateManager.TrackPod(&api.Pod{
		Namespace: "ns1",
		Name:      "pod1",
		Status:    "ImagePullBackOff",
	}, "", "")

	relatedAlerts, _ := store.RelatedAlerts("ns1/pod1")
	assert.Equal(t, 1, len(relatedAlerts))
	assert.Equal(t, "imagePullBackOffThreshold", relatedAlerts[0].Type)
}

// func TestTrackEvents(t *testing.T) {
// 	store := store.NewTest(encryptionKey, encryptionKeyNew)
// 	defer func() {
// 		store.Close()
// 	}()

// 	dummyNotificationsManager := notifications.NewDummyManager()
// 	event1 := api.Event{Namespace: "ns1", Name: "pod1"}
// 	event2 := api.Event{Namespace: "ns1", Name: "pod2"}
// 	events := []api.Event{event1, event2}

// 	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
// 	p.TrackEvents(events)

// 	expectedEvents := []model.KubeEvent{
// 		{Name: "ns1/pod1"},
// 		{Name: "ns1/pod2"},
// 	}
// 	for _, event := range expectedEvents {
// 		e, _ := store.KubeEvent(event.Name)

// 		assert.Equal(t, event.Name, e.Name)
// 	}

// 	expectedAlerts := []model.Alert{
// 		{Name: "ns1/pod1", Type: "event", Status: "Pending"},
// 		{Name: "ns1/pod2", Type: "event", Status: "Pending"},
// 	}
// 	for _, alert := range expectedAlerts {
// 		a, _ := store.Alert(alert.Name, alert.Type)

// 		assert.Equal(t, alert.Status, a.Status)
// 	}
// }

// func TestSetFiringState(t *testing.T) {
// 	store := store.NewTest(encryptionKey, encryptionKeyNew)
// 	defer func() {
// 		store.Close()
// 	}()

// 	currentTime := time.Now()
// 	dummyNotificationsManager := notifications.NewDummyManager()
// 	a := NewAlertStateManager(dummyNotificationsManager, *store, 2)
// 	alert1 := model.Alert{Name: "n/p1", Type: "pod", Status: "Pending", LastStateChange: currentTime.Add(-1 * time.Minute).Unix()}
// 	alert2 := model.Alert{Name: "n/p2", Type: "pod", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix()}
// 	alert3 := model.Alert{Name: "n/p1", Type: "event", Status: "Pending", LastStateChange: currentTime.Add(-4 * time.Minute).Unix(), Count: 5}
// 	alert4 := model.Alert{Name: "n/p2", Type: "event", Status: "Pending", LastStateChange: currentTime.Add(-3 * time.Minute).Unix(), Count: 7}
// 	alerts := []model.Alert{alert1, alert2, alert3, alert4}

// 	for _, alert := range alerts {
// 		store.SaveOrUpdateAlert(&alert)
// 	}

// 	var thresholds []threshold
// 	for _, alert := range alerts {
// 		thresholds = append(thresholds, ToThreshold(&alert, a.waitTime, 6, 1))
// 	}

// 	err := a.setFiringState(thresholds)
// 	assert.Nil(t, err)

// 	expected := []model.Alert{
// 		{Name: "n/p1", Type: "pod", Status: "Pending"},
// 		{Name: "n/p2", Type: "pod", Status: "Firing"},
// 		{Name: "n/p1", Type: "event", Status: "Pending"},
// 		{Name: "n/p2", Type: "event", Status: "Firing"},
// 	}
// 	for _, alert := range expected {
// 		a, _ := store.Alert(alert.Name, alert.Type)
// 		assert.Equal(t, alert.Status, a.Status)
// 	}
// }

// func TestPodFailedMessage(t *testing.T) {
// 	msgPodFailed := notifications.AlertMessage{
// 		Alert: model.Alert{
// 			Type:       "pod",
// 			Name:       "ns1/pod1",
// 			StatusDesc: "Container failed",
// 		},
// 	}

// 	discordMsg, err := msgPodFailed.AsDiscordMessage()
// 	assert.Nil(t, err)

// 	assert.Contains(t, "pod ns1/pod1 failed", discordMsg.Text)

// 	slackMsg, err := msgPodFailed.AsSlackMessage()
// 	assert.Nil(t, err)

// 	assert.Contains(t, "pod ns1/pod1 failed", slackMsg.Text)
// }

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
