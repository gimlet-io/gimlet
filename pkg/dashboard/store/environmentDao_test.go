package store

import (
	"testing"
	"time"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentCreateAndGetAll(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	environmentStaging := model.Environment{
		Name: "staging",
	}

	environmentProd := model.Environment{
		Name: "prod",
	}

	environmentEphemeral := model.Environment{
		Name:      "ephemeral",
		Ephemeral: true,
		Expiry:    time.Now().Unix(),
	}

	errCreateEnvStaging := s.CreateEnvironment(&environmentStaging)
	assert.Nil(t, errCreateEnvStaging)

	errCreateEnvProd := s.CreateEnvironment(&environmentProd)
	assert.Nil(t, errCreateEnvProd)

	errCreateEnvEphemeral := s.CreateEnvironment(&environmentEphemeral)
	assert.Nil(t, errCreateEnvEphemeral)

	envArray, err := s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, 3, len(envArray))
	assert.Equal(t, environmentProd.Name, envArray[1].Name)
	assert.Equal(t, environmentStaging.Name, envArray[2].Name)
	assert.Equal(t, environmentEphemeral.Ephemeral, envArray[0].Ephemeral)

	staging := envArray[2]
	staging.InfraRepo = "my-custom-repo"
	err = s.UpdateEnvironment(staging)
	if err != nil {
		t.Errorf("Cannot update environment: %s", err)
	}

	envArray, err = s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, "my-custom-repo", envArray[2].InfraRepo)
}

func TestEnvironmentDelete(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	environmentStaging := model.Environment{
		Name: "staging",
	}

	errCreateEnvStaging := s.CreateEnvironment(&environmentStaging)
	assert.Nil(t, errCreateEnvStaging)

	s.DeleteEnvironment("staging")

	data, _ := s.GetEnvironments()

	assert.Equal(t, 0, len(data))
}
