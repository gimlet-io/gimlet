package main

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/stretchr/testify/assert"
)

func TestParseEnvs(t *testing.T) {
	input := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=production&repoPerEnv=true&infraRepo=gitops-infra2&appsRepo=gitops-apps2"
	envs, err := parseEnvs(input)
	if err != nil {
		t.Errorf("Cannot parse environments: %s", err)
	}

	assert.Equal(t, 2, len(envs))
	assert.Equal(t, "staging", envs[0].Name)
	assert.Equal(t, false, envs[0].RepoPerEnv)
	assert.Equal(t, "gitops-infra", envs[0].InfraRepo)
	assert.Equal(t, "gitops-apps", envs[0].AppsRepo)
}

func TestParseEnvs_Empty(t *testing.T) {
	input := ""
	_, err := parseEnvs(input)
	assert.Nil(t, err)
}

func TestBootstrapEnvs(t *testing.T) {
	s := store.NewTest()
	defer func() {
		s.Close()
	}()

	validInput := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=live&repoPerEnv=true&infraRepo=extra-gitops-infra&appsRepo=extra-gitops-apps"
	inputWithoutName := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;repoPerEnv=true&infraRepo=extra-gitops-infra&appsRepo=extra-gitops-apps"
	inputWithoutInfraRepo := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=ready&repoPerEnv=true&appsRepo=extra-gitops-apps"
	inputWithoutAppsRepo := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=fire&repoPerEnv=true&infraRepo=extra-gitops-infra&"
	expectedErrorMessage := "name, infraRepo, and appsRepo are mandatory for environments"

	environmentStaging := model.Environment{
		Name:       "staging",
		RepoPerEnv: true,
		InfraRepo:  "infra-repo",
		AppsRepo:   "apps-repo",
	}

	environmentProduction := model.Environment{
		Name:       "production",
		RepoPerEnv: false,
		InfraRepo:  "infra-repo",
		AppsRepo:   "apps-repo",
	}

	errCreateEnvStaging := s.CreateEnvironment(&environmentStaging)
	assert.Nil(t, errCreateEnvStaging)

	errCreateEnvProd := s.CreateEnvironment(&environmentProduction)
	assert.Nil(t, errCreateEnvProd)

	err := bootstrapEnvs(validInput, s)
	if err != nil {
		t.Errorf("Cannot bootstrap environments: %s", err)
	}

	envs, err := s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, 3, len(envs))
	assert.Equal(t, "live", envs[0].Name)
	assert.Equal(t, true, envs[0].RepoPerEnv)
	assert.Equal(t, "extra-gitops-infra", envs[0].InfraRepo)
	assert.Equal(t, "extra-gitops-apps", envs[0].AppsRepo)

	assert.EqualError(t, bootstrapEnvs(inputWithoutName, s), expectedErrorMessage)
	assert.EqualError(t, bootstrapEnvs(inputWithoutInfraRepo, s), expectedErrorMessage)
	assert.EqualError(t, bootstrapEnvs(inputWithoutAppsRepo, s), expectedErrorMessage)
}
