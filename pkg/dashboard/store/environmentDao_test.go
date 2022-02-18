package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentCreateAndGetAll(t *testing.T) {
	s := NewTest()
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
	assert.Equal(t, environmentStaging.Name, envArray[0].Name)
	assert.Equal(t, environmentProd.Name, envArray[1].Name)
}

func TestEnvironmentDelete(t *testing.T) {
	s := NewTest()
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
