package notifications

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
)

func TestSavePodState(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod := model.Pod{Deployment: "n/p", Status: "ErrImagePull"}

	err := store.SaveOrUpdatePod(&pod)
	assert.Nil(t, err)

	podFromDb, err := store.Pod(pod.Deployment)
	assert.Nil(t, err)

	assert.Equal(t, podFromDb.Status, pod.Status)

	updatedPod := model.Pod{Deployment: "n/p", Status: "Running"}

	err = store.SaveOrUpdatePod(&updatedPod)
	assert.Nil(t, err)

	updatedPodFromDb, err := store.Pod(pod.Deployment)
	assert.Nil(t, err)

	assert.Equal(t, updatedPod.Status, updatedPodFromDb.Status)
}

func TestGetPodFromDB(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod1 := model.Pod{Deployment: "n/p", Status: "ErrImagePull"}
	pod2 := model.Pod{Deployment: "n/p2", Status: "Running"}

	err := store.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	err = store.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	podFromDb, err := store.Pod(pod1.Deployment)
	assert.Nil(t, err)

	assert.Equal(t, pod1.Status, podFromDb.Status)
}

func TestTrackStates(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pod3 := api.Pod{Namespace: "ns2", Name: "pod3", Status: "Pending"}
	pods := []api.Pod{pod1, pod2, pod3}

	p := NewPodStateManager(NewDummyManager(), 2)
	p.trackStates(pods, *store)

	expected := []model.Pod{
		{Deployment: "ns1/pod1", Status: "Running"},
		{Deployment: "ns1/pod2", Status: "PodFailed"},
		{Deployment: "ns2/pod3", Status: "Pending"},
	}
	for _, pod := range expected {
		p, err := store.Pod(pod.Deployment)
		assert.Nil(t, err)

		assert.Equal(t, p.Status, pod.Status)
	}
}
