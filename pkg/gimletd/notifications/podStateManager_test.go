package notifications

import (
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
)

func TestSavePods(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod1 := model.Pod{Name: "n/p1", Status: "ErrImagePull"}

	err := s.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	pods, err := s.Pods()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pods))

	pod2 := model.Pod{Name: "n/p2", Status: "Running"}

	err = s.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	pods, err = s.Pods()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pods))
}

func TestSavePodState(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod := model.Pod{Name: "n/p", Status: "ErrImagePull"}

	err := store.SaveOrUpdatePod(&pod)
	assert.Nil(t, err)

	podFromDb, err := store.Pod(pod.Name)
	assert.Nil(t, err)

	assert.Equal(t, podFromDb.Status, pod.Status)

	updatedPod := model.Pod{Name: "n/p", Status: "Running"}

	err = store.SaveOrUpdatePod(&updatedPod)
	assert.Nil(t, err)

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

	err := store.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	err = store.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	podFromDb, err := store.Pod(pod1.Name)
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

	p := NewPodStateManager(NewDummyManager(), *store, 2)
	p.trackStates(pods)

	expected := []model.Pod{
		{Name: "ns1/pod1", Status: "Running"},
		{Name: "ns1/pod2", Status: "PodFailed"},
		{Name: "ns2/pod3", Status: "Pending"},
	}
	for _, pod := range expected {
		p, err := store.Pod(pod.Name)
		assert.Nil(t, err)

		assert.Equal(t, p.Status, pod.Status)
	}
}

func TestPodAlertStates(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	pod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Running"}
	pod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "PodFailed"}
	pods := []api.Pod{pod1, pod2}

	p := NewPodStateManager(NewDummyManager(), *store, 2)
	p.trackStates(pods)

	expected := []model.Pod{
		{Name: "ns1/pod1", Status: "Running", AlertState: "OK"},
		{Name: "ns1/pod2", Status: "PodFailed", AlertState: "Pending"},
	}
	for _, pod := range expected {
		p, err := store.Pod(pod.Name)
		assert.Nil(t, err)

		assert.Equal(t, p.AlertState, pod.AlertState)
	}

	updatedPod1 := api.Pod{Namespace: "ns1", Name: "pod1", Status: "Error"}
	p.trackStates([]api.Pod{updatedPod1})

	updatedPod1FromDb, err := store.Pod(fmt.Sprintf("%s/%s", updatedPod1.Namespace, updatedPod1.Name))
	assert.Nil(t, err)

	assert.Equal(t, updatedPod1FromDb.AlertState, "Pending")

	updatedPod2 := api.Pod{Namespace: "ns1", Name: "pod2", Status: "Running"}
	p.trackStates([]api.Pod{updatedPod2})

	updatedPod2FromDb, err := store.Pod(fmt.Sprintf("%s/%s", updatedPod2.Namespace, updatedPod2.Name))
	assert.Nil(t, err)

	assert.Equal(t, updatedPod2FromDb.AlertState, "OK")
}

func TestPendingStateExpired(t *testing.T) {
	p := NewPodStateManager(NewDummyManager(), *store.NewTest(), 2)

	isPending1 := p.isPendingStateExpired(time.Now().Add(-time.Minute * 2).Unix())
	assert.Equal(t, true, isPending1)

	isPending2 := p.isPendingStateExpired(time.Now().Add(-time.Minute * 1).Unix())
	assert.Equal(t, false, isPending2)
}
