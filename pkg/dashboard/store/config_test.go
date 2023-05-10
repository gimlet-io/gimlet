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

	configsValueFromDb, err := s.GetConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(configsValueFromDb))
}
