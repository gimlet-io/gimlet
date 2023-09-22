package store

import (
	"testing"

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

	err = s.UpdateAlertState(alert.ID, model.FIRING)
	assert.Nil(t, err)

	alerts, err := s.AlertsByState(model.FIRING)
	assert.Nil(t, err)

	assert.Equal(t, model.FIRING, alerts[0].Status)
}
