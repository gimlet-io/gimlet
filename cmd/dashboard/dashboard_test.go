package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
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
	encryptionKey := "the-key-has-to-be-32-bytes-long!"
	encryptionKeyNew := ""
	s := store.NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	validInput := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=live&repoPerEnv=true&infraRepo=extra-gitops-infra&appsRepo=extra-gitops-apps"
	inputWithoutName := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;repoPerEnv=true&infraRepo=extra-gitops-infra&appsRepo=extra-gitops-apps"
	inputWithoutInfraRepo := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=ready&repoPerEnv=true&appsRepo=extra-gitops-apps"
	inputWithoutAppsRepo := "name=staging&repoPerEnv=false&infraRepo=gitops-infra&appsRepo=gitops-apps;name=fire&repoPerEnv=true&infraRepo=extra-gitops-infra&"
	expectedErrorMessage := "name, infraRepo, and appsRepo are mandatory for environments"
	fromGimletdConfig := "env=staging2&repoPerEnv=false&gitopsRepo=gitops-staging-infra&deployKeyPath=/deploykey/staging.key"

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

	err := bootstrapEnvs(validInput, s, fromGimletdConfig)
	if err != nil {
		t.Errorf("Cannot bootstrap environments: %s", err)
	}

	envs, err := s.GetEnvironments()
	if err != nil {
		t.Errorf("Cannot get environments: %s", err)
	}

	assert.Equal(t, 4, len(envs))
	assert.Equal(t, "live", envs[0].Name)
	assert.Equal(t, true, envs[0].RepoPerEnv)
	assert.Equal(t, "extra-gitops-infra", envs[0].InfraRepo)
	assert.Equal(t, "extra-gitops-apps", envs[0].AppsRepo)

	assert.EqualError(t, bootstrapEnvs(inputWithoutName, s, ""), expectedErrorMessage)
	assert.EqualError(t, bootstrapEnvs(inputWithoutInfraRepo, s, ""), expectedErrorMessage)
	assert.EqualError(t, bootstrapEnvs(inputWithoutAppsRepo, s, ""), expectedErrorMessage)
}

func TestJWTexpiryWithExpiredToken(t *testing.T) {
	secret := "mySecretString"
	subtractTwelveHours, _ := time.ParseDuration("-12h")
	exp := time.Now().Add(subtractTwelveHours).Unix()
	expiredJWT := token.New("sess", "test")
	expiredJWTStr, err := expiredJWT.SignExpires(secret, exp)
	if err != nil {
		t.Errorf("Cannot sign token expiration time: %s", err)
	}

	_, err = token.Parse(expiredJWTStr, func(t *token.Token) (string, error) {
		return secret, nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Token must return expired: %s", err)
		}
	} else {
		t.Error("Token must return expired")
	}
}

func TestJWTexpiryWithValidToken(t *testing.T) {
	secret := "mySecretString"
	twelveHours, _ := time.ParseDuration("12h")
	exp := time.Now().Add(twelveHours).Unix()
	validJWT := token.New("sess", "test")
	validJWTStr, err := validJWT.SignExpires(secret, exp)
	if err != nil {
		t.Errorf("Cannot sign token expiration time: %s", err)
	}

	_, err = token.Parse(validJWTStr, func(t *token.Token) (string, error) {
		return secret, nil
	})
	if err != nil {
		t.Errorf("Token must return valid: %s", err)
	}
}

func TestParseChannelMapping(t *testing.T) {
	config := &config.Config{
		Notifications: config.Notifications{
			ChannelMapping: "staging=my-team,prod=another-team",
		},
	}

	testChannelMap := parseChannelMap(config)

	assertEqual(t, testChannelMap["staging"], "my-team")
	assertEqual(t, testChannelMap["prod"], "another-team")
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestParseGitopsRepos(t *testing.T) {
	validInput := "env=staging&repoPerEnv=false&gitopsRepo=gitops-staging-infra&deployKeyPath=/deploykey/staging.key;env=production&repoPerEnv=true&gitopsRepo=gitops-production-infra&deployKeyPath=/deploykey/production.key"

	gitopsRepos, err := parseGitopsRepos(validInput)
	if err != nil {
		t.Errorf("Cannot parse gitopsRepos")
	}

	assert.Equal(t, 2, len(gitopsRepos))
	assert.Equal(t, "staging", gitopsRepos[0].Name)
	assert.Equal(t, false, gitopsRepos[0].RepoPerEnv)
	assert.Equal(t, "gitops-staging-infra", gitopsRepos[0].AppsRepo)
	assert.Equal(t, "production", gitopsRepos[1].Name)
	assert.Equal(t, true, gitopsRepos[1].RepoPerEnv)
	assert.Equal(t, "gitops-production-infra", gitopsRepos[1].AppsRepo)
}
