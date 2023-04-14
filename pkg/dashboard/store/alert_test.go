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
		ObjectType: "pod",
		ObjectName: "default/pod1",
		Status:     "Firing",
	}

	err := s.CreateAlert(&alert)
	assert.Nil(t, err)

	alerts, err := s.FiringAlerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alerts))

	a, err := s.Alerts(alert.ObjectName, alert.ObjectType)
	assert.Nil(t, err)
	assert.Equal(t, alert.ObjectName, a[0].ObjectName)
}

func TestGetPendingAlerts(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	alert1 := model.Alert{
		ObjectType: "pod",
		ObjectName: "default/pod1",
		Status:     "Pending",
	}

	alert2 := model.Alert{
		ObjectName: "default/pod2",
		Status:     "Firing",
	}

	s.CreateAlert(&alert1)
	s.CreateAlert(&alert2)

	pendingAlerts, err := s.PendingAlerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pendingAlerts))
}
