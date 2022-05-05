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
	assert.Equal(t, "staging", gitopsRepos["staging"].Env)
	assert.Equal(t, false, gitopsRepos["staging"].RepoPerEnv)
	assert.Equal(t, "gitops-staging-infra", gitopsRepos["staging"].GitopsRepo)
	assert.Equal(t, "/deploykey/staging.key", gitopsRepos["staging"].DeployKeyPath)
	assert.Equal(t, "production", gitopsRepos["production"].Env)
	assert.Equal(t, true, gitopsRepos["production"].RepoPerEnv)
	assert.Equal(t, "gitops-production-infra", gitopsRepos["production"].GitopsRepo)
	assert.Equal(t, "/deploykey/production.key", gitopsRepos["production"].DeployKeyPath)
}
