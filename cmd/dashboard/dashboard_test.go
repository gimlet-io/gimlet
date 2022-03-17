package main

import (
	"testing"

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
}
