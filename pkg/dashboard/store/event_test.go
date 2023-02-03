package store

import (
	"database/sql"
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
		Name:  "default/pod1",
		Count: 1,
	}

	err := s.SaveOrUpdateEvent(&event)
	assert.Nil(t, err)

	e, err := s.Event(event.Name)
	assert.Nil(t, err)
	assert.Equal(t, event.Name, e.Name)

	s.SaveOrUpdateEvent(&model.Event{
		Name:  "default/pod1",
		Count: 2,
	})

	e, _ = s.Event(event.Name)
	assert.Equal(t, int32(2), e.Count)

	err = s.DeleteEvent(event.Name)
	assert.Nil(t, err)

	_, err = s.Event(event.Name)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetPendingEvents(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	s.SaveOrUpdateEvent(&model.Event{
		Name:       "default/pod2",
		AlertState: "Pending",
	})

	s.SaveOrUpdateEvent(&model.Event{
		Name:       "default/pod2",
		AlertState: "OK",
	})

	e, err := s.PendingEvents()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(e))
}
