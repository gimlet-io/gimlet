package store

import (
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func Test_RelatedAlerts(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	alerts, err := s.RelatedAlerts("pod")
	assert.Nil(t, err)
	assert.NotNil(t, alerts)
}

func Test_SetFiringAlertState(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	alert, err := s.CreateAlert(&model.Alert{
		ObjectName:     "pod",
		Type:           "Failed",
		DeploymentName: "deployment",
		Status:         model.PENDING,
	})
	assert.Nil(t, err)

	alert.SetFiring()
	err = s.UpdateAlertState(alert)
	assert.Nil(t, err)

	alerts, err := s.AlertsByState(model.FIRING)
	assert.Nil(t, err)

	assert.Equal(t, model.FIRING, alerts[0].Status)
}

func Test_GetAlerts(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	s.CreateAlert(&model.Alert{
		ObjectName: "pod-new",
		FiredAt:    time.Now().Add(-1 * time.Hour * 1).Unix(),
	})
	s.CreateAlert(&model.Alert{
		ObjectName: "pod-old",
		FiredAt:    time.Now().Add(-1 * time.Hour * 30).Unix(),
	})

	alerts, err := s.Alerts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alerts))
	assert.Equal(t, "pod-new", alerts[0].ObjectName)
}
