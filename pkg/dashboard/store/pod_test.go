package store

import (
	"database/sql"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestPodCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	pod := model.Pod{
		Name:   "default/pod1",
		Status: "Running",
	}

	err := s.SaveOrUpdatePod(&pod)
	assert.Nil(t, err)

	p, err := s.Pod(pod.Name)
	assert.Nil(t, err)
	assert.Equal(t, pod.Name, p.Name)

	err = s.DeletePod(pod.Name)
	assert.Nil(t, err)

	_, err = s.Pod(pod.Name)
	assert.Equal(t, sql.ErrNoRows, err)
}
