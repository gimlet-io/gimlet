package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestEventCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	event := model.Event{
		Name: "default/pod1",
	}

	err := s.SaveOrUpdateEvent(&event)
	assert.Nil(t, err)

	a, err := s.Event(event.Name)
	assert.Nil(t, err)
	assert.Equal(t, event.Name, a.Name)
}
