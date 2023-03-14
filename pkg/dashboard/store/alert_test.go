package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestAlertCRUD(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	alert := model.Alert{
		Type:   "pod",
		Name:   "default/pod1",
		Status: "Firing",
	}

	err := s.SaveOrUpdateAlert(&alert)
	assert.Nil(t, err)

	alerts, err := s.FiringAlerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alerts))

	a, err := s.Alert(alert.Name, alert.Type)
	assert.Nil(t, err)
	assert.Equal(t, alert.Name, a.Name)
}

func TestGetPendingAlerts(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
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

	s.SaveOrUpdateAlert(&alert1)
	s.SaveOrUpdateAlert(&alert2)

	pendingAlerts, err := s.PendingAlerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pendingAlerts))
}
