package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestConfigCreateAndRead(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	config := model.Config{
		Key:   "ORG",
		Value: "gimlet-io",
	}

	err := s.SaveConfig(&config)
	assert.Nil(t, err)

	configValueFromDb, err := s.GetConfigValue("ORG")
	assert.Nil(t, err)
	assert.Equal(t, config.Value, configValueFromDb)

	err = s.SaveConfig(&config)
	assert.Error(t, err, "Expected SaveConfig to return an error for existing key")
}
