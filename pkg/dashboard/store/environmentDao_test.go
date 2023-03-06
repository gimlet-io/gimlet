package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
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

	errCreateEnvStaging := s.CreateEnvironment(&environmentStaging)
	assert.Nil(t, errCreateEnvStaging)

	errCreateEnvProd := s.CreateEnvironment(&environmentProd)
	assert.Nil(t, errCreateEnvProd)

	envArray, err := s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, 2, len(envArray))
	assert.Equal(t, environmentProd.Name, envArray[0].Name)
	assert.Equal(t, environmentStaging.Name, envArray[1].Name)

	staging := envArray[1]
	staging.InfraRepo = "my-custom-repo"
	err = s.UpdateEnvironment(staging)
	if err != nil {
		t.Errorf("Cannot update environment: %s", err)
	}

	envArray, err = s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, "my-custom-repo", envArray[1].InfraRepo)
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
