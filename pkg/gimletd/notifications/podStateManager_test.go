package notifications

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
)

func TestSavePodState(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod := model.Pod{Name: "p", Namespace: "n", Status: "ErrImagePull"}

	err := s.SaveOrUpdatePod(&pod)
	assert.Nil(t, err)

	podFromDb, err := s.Pod(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	assert.Nil(t, err)

	assert.Equal(t, podFromDb.Status, pod.Status)

	updatedPod := model.Pod{Name: "p", Namespace: "n", Status: "Running"}

	err = s.SaveOrUpdatePod(&updatedPod)
	assert.Nil(t, err)

	updatedPodFromDb, err := s.Pod(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	assert.Nil(t, err)

	assert.Equal(t, updatedPod.Status, updatedPodFromDb.Status)
}

func TestGetPodFromDB(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	pod1 := model.Pod{Name: "p", Namespace: "n", Status: "ErrImagePull"}
	pod2 := model.Pod{Name: "p2", Namespace: "n", Status: "Running"}

	err := s.SaveOrUpdatePod(&pod1)
	assert.Nil(t, err)

	err = s.SaveOrUpdatePod(&pod2)
	assert.Nil(t, err)

	podFromDb, err := s.Pod(fmt.Sprintf("%s/%s", pod1.Namespace, pod1.Name))
	assert.Nil(t, err)

	assert.Equal(t, pod1.Status, podFromDb.Status)
}

func TestPodStateManagerTrackStates(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	dummyPodStateManager := NewPodStateManager(NewDummyManager())

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
