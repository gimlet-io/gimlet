package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestAlertCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	alert := model.Alert{
		Type: "pod",
		Name: "default/pod1",
	}

	err := s.SaveAlert(&alert)
	assert.Nil(t, err)

	alerts, err := s.Alerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alerts))

	a, err := s.Alert(alert.Name, alert.Type)
	assert.Nil(t, err)
	assert.Equal(t, alert.Name, a.Name)

	alerts, _ = s.Alerts()
	assert.Equal(t, 0, len(alerts))

}

//TODO write test for pending alerts
