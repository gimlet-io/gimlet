package main

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"gotest.tools/assert"
)

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
	assert.Equal(t, "staging", gitopsRepos[0].Env)
	assert.Equal(t, false, gitopsRepos[0].RepoPerEnv)
	assert.Equal(t, "gitops-staging-infra", gitopsRepos[0].GitopsRepo)
	assert.Equal(t, "/deploykey/staging.key", gitopsRepos[0].DeployKeyPath)
	assert.Equal(t, "production", gitopsRepos[1].Env)
	assert.Equal(t, true, gitopsRepos[1].RepoPerEnv)
	assert.Equal(t, "gitops-production-infra", gitopsRepos[1].GitopsRepo)
	assert.Equal(t, "/deploykey/production.key", gitopsRepos[1].DeployKeyPath)
}

