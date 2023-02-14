package store

import (
	"database/sql"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestKubeEventCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	event := model.KubeEvent{
		Name: "default/pod1",
	}

	err := s.SaveOrUpdateKubeEvent(&event)
	assert.Nil(t, err)

	e, err := s.KubeEvent(event.Name)
	assert.Nil(t, err)
	assert.Equal(t, event.Name, e.Name)

	err = s.DeleteEvent(event.Name)
	assert.Nil(t, err)

	_, err = s.KubeEvent(event.Name)
	assert.Equal(t, sql.ErrNoRows, err)
}
