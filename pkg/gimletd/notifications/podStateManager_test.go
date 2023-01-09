package notifications

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gimlet-io/gimletd/notifications"
)

func TestSavePodState(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod := model.Pod{}

	err := s.SaveOrUpdatePod(&pod)
	assert.Nil(t, err)

	pods, err := s.Pods()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pods))
	assert.Equal(t, pod.Status, pods[0].Status)
}

func TestGetPodFromDB(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod1 := model.Pod{}
	pod2 := model.Pod{}

	err := s.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	err = s.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	pod, err := s.Pod(fmt.Sprintf("%s/%s", pod1.Namespace, pod1.Name))
	assert.Nil(t, err)

	assert.Equal(t, pod1.Status, pod.Status)
}

func TestPodStateManagerTrackStates(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	dummyPodStateManager := NewPodStateManager(notifications.NewDummyManager())

	dummyPods := []api.Pod{
		{Name: "p", Namespace: "n", Status: "ErrImagePull"},
		{Name: "p2", Namespace: "n", Status: "Running"},
		{Name: "p3", Namespace: "n", Status: "Error"},
	}

	pod1 := model.Pod{Name: "p", Namespace: "n", Status: "ErrImagePull"}
	pod2 := model.Pod{Name: "p2", Namespace: "n", Status: "Running"}

	err := s.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	err = s.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	dummyPodStateManager.trackStates(dummyPods, *s)
}
