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
}

func TestGetPendingAlerts(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	alert1 := model.Alert{
		Type:   "pod",
		Name:   "default/pod1",
		Status: "Pending",
	}

	alert2 := model.Alert{
		Name:   "default/pod2",
		Status: "Firing",
	}

	s.SaveAlert(&alert1)
	s.SaveAlert(&alert2)

	pendingAlerts, err := s.PendingAlertsByType(alert1.Type)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pendingAlerts))
}

func TestSetFiringStatusForAlert(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	alert1 := model.Alert{
		Type:   "pod",
		Name:   "default/pod1",
		Status: "Pending",
	}

	s.SaveAlert(&alert1)

	err := s.SetFiringStatusForAlert(alert1.Name, alert1.Type)
	assert.Nil(t, err)

	a, _ := s.Alert(alert1.Name, alert1.Type)
	assert.Equal(t, "Firing", a.Status)
}
