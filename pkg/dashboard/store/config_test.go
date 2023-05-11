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
		Key:   GithubOrg,
		Value: "gimlet-io",
	}

	err := s.SaveConfig(&config)
	assert.Nil(t, err)

	configFromDb, err := s.GetConfig(GithubOrg)
	assert.Nil(t, err)
	assert.Equal(t, config.Value, configFromDb.Value)
}
