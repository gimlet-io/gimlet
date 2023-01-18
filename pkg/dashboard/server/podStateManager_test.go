package server

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
)

func TestGetPendingPods(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod1 := model.Pod{Name: "n/p1", AlertState: "OK"}

	err := s.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	pods, err := s.PendingPods()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(pods))

	pod2 := model.Pod{Name: "n/p2", AlertState: "Pending"}

	err = s.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	pods, err = s.PendingPods()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pods))
}

func TestDeletePod(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod1 := model.Pod{Name: "n/p1", AlertState: "OK"}

	s.SaveOrUpdatePod(&pod1)

	err := s.DeletePod(pod1.Name)
	assert.Nil(t, err)

	_, err = s.Pod(pod1.Name)

	assert.Equal(t, err, sql.ErrNoRows)
}

func TestSavePodState(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod := model.Pod{Name: "n/p", Status: "ErrImagePull"}

	store.SaveOrUpdatePod(&pod)

	podFromDb, err := store.Pod(pod.Name)
	assert.Nil(t, err)

	assert.Equal(t, podFromDb.Status, pod.Status)

	updatedPod := model.Pod{Name: "n/p", Status: "Running"}

	store.SaveOrUpdatePod(&updatedPod)

	updatedPodFromDb, err := store.Pod(updatedPod.Name)
	assert.Nil(t, err)

	assert.Equal(t, updatedPod.Status, updatedPodFromDb.Status)
}

func TestGetPodFromDB(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod1 := model.Pod{Name: "n/p", Status: "ErrImagePull"}
	pod2 := model.Pod{Name: "n/p2", Status: "Running"}

	store.SaveOrUpdatePod(&pod1)

	store.SaveOrUpdatePod(&pod2)

	podFromDb, _ := store.Pod(pod1.Name)

	assert.Equal(t, pod1.Status, podFromDb.Status)
}

func TestTrackStates(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pod3 := api.Pod{Namespace: "ns2", Name: "pod3", Status: "Pending"}
	pods := []*api.Pod{&pod1, &pod2, &pod3}

	p := NewPodStateManager(dummyNotificationsManager, *store, 2)
	p.trackStates(pods)

	expected := []model.Pod{
		{Name: "ns1/pod1", Status: "Running"},
		{Name: "ns1/pod2", Status: "PodFailed"},
		{Name: "ns2/pod3", Status: "Pending"},
	}
	for _, pod := range expected {
		p, _ := store.Pod(pod.Name)

		assert.Equal(t, p.Status, pod.Status)
	}
}

func TestPodAlertStates(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pods := []*api.Pod{&pod1, &pod2}

	p := NewPodStateManager(dummyNotificationsManager, *store, 2)
	p.trackStates(pods)

	expected := []model.Pod{
		{Name: "ns1/pod1", AlertState: "OK"},
		{Name: "ns1/pod2", AlertState: "Pending"},
	}
	for _, pod := range expected {
		p, _ := store.Pod(pod.Name)

		assert.Equal(t, p.AlertState, pod.AlertState)
	}

	updatedPod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Error"}
	p.trackStates([]*api.Pod{&updatedPod1})

	updatedPod1FromDb, _ := store.Pod(fmt.Sprintf("%s/%s", updatedPod1.Namespace, updatedPod1.Name))

	assert.Equal(t, updatedPod1FromDb.AlertState, "Pending")

	updatedPod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "Running"}
	p.trackStates([]*api.Pod{&updatedPod2})

	updatedPod2Name := fmt.Sprintf("%s/%s", updatedPod2.Namespace, updatedPod2.Name)
	updatedPod2FromDb, _ := store.Pod(updatedPod2Name)

	assert.Equal(t, updatedPod2FromDb.AlertState, "OK")
}

func TestSetPodAlertStatesToFiring(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	currentTime := time.Now()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := model.Pod{Name: "ns1/pod1", Status: "Error", AlertState: "Pending", AlertStateTimestamp: currentTime.Add(-1 * time.Minute).Unix()}
	pod2 := model.Pod{Name: "ns1/pod2", Status: "PodFailed", AlertState: "Pending", AlertStateTimestamp: currentTime.Add(-3 * time.Minute).Unix()}

	store.SaveOrUpdatePod(&pod1)
	store.SaveOrUpdatePod(&pod2)

	p := NewPodStateManager(dummyNotificationsManager, *store, 2)

	p.setFiringState()

	expected := []model.Pod{
		{Name: "ns1/pod1", AlertState: "Pending"},
		{Name: "ns1/pod2", AlertState: "Firing"},
	}
	for _, pod := range expected {
		p, _ := store.Pod(pod.Name)
		assert.Equal(t, p.AlertState, pod.AlertState)
	}
}

func TestPodAlertStateTimestampOverwrite(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	oneMinuteAgo := time.Now().Add(-1 * time.Minute).Unix()

	dummyNotificationsManager := notifications.NewDummyManager()
	pod1 := model.Pod{Name: "ns1/pod1", Status: "Error", AlertState: "Pending", AlertStateTimestamp: oneMinuteAgo}
	pod2 := model.Pod{Name: "ns1/pod2", Status: "Running", AlertState: "OK", AlertStateTimestamp: oneMinuteAgo}

	store.SaveOrUpdatePod(&pod1)
	store.SaveOrUpdatePod(&pod2)

	p := NewPodStateManager(dummyNotificationsManager, *store, 2)

	trackablePod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Error"}
	trackablePod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "Running"}

	p.trackStates([]*api.Pod{&trackablePod1, &trackablePod2})

	expected := []model.Pod{
		{Name: "ns1/pod1", AlertStateTimestamp: oneMinuteAgo},
		{Name: "ns1/pod2", AlertStateTimestamp: oneMinuteAgo},
	}
	for _, pod := range expected {
		p, _ := store.Pod(pod.Name)
		assert.Equal(t, p.AlertStateTimestamp, pod.AlertStateTimestamp)
	}

	trackablePod1 = api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	trackablePod2 = api.Pod{Namespace: "ns1", Name: "pod2", Status: "Error"}

	p.trackStates([]*api.Pod{&trackablePod1, &trackablePod2})

	expected = []model.Pod{
		{Name: "ns1/pod1", AlertStateTimestamp: oneMinuteAgo},
		{Name: "ns1/pod2", AlertStateTimestamp: oneMinuteAgo},
	}
	for _, pod := range expected {
		p, _ := store.Pod(pod.Name)
		assert.NotEqual(t, p.AlertStateTimestamp, pod.AlertStateTimestamp)
	}
}

func TestPodFailedMessage(t *testing.T) {
	msgPodFailed := notifications.PodFailedMessage{
		Pod: model.Pod{
			Name:       "ns1/pod1",
			Status:     "Error",
			StatusDesc: "Container failed",
		},
	}

	discordMsg, err := msgPodFailed.AsDiscordMessage()
	assert.Nil(t, err)

	assert.Contains(t, discordMsg.Text, "ns1/pod1 failed with status Error")

	slackMsg, err := msgPodFailed.AsSlackMessage()
	assert.Nil(t, err)

	assert.Contains(t, slackMsg.Text, "ns1/pod1 failed with status Error")
}

func TestNotificationSending(t *testing.T) {
	t.Skip("Skipping notification sending")
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	notificationsManager := notifications.NewManager()

	notificationsManager.AddProvider(&notifications.DiscordProvider{
		Token:     "",
		ChannelID: "",
	})

	go notificationsManager.Run()

	p := NewPodStateManager(notificationsManager, *store, 2)

	currentTime := time.Now()
	pod1 := model.Pod{Name: "ns1/pod1", Status: "Error", StatusDesc: "Back-off pulling image", AlertState: "Pending", AlertStateTimestamp: currentTime.Add(-1 * time.Minute).Unix()}
	pod2 := model.Pod{Name: "ns1/pod2", Status: "ErrImagePull", StatusDesc: "Back-off pulling image", AlertState: "Pending", AlertStateTimestamp: currentTime.Add(-3 * time.Minute).Unix()}

	store.SaveOrUpdatePod(&pod1)
	store.SaveOrUpdatePod(&pod2)

	p.setFiringState()

	time.Sleep(5 * time.Second)
}
