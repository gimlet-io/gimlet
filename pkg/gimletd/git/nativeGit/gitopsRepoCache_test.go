package nativeGit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitopsRepos(t *testing.T) {
	validInput := "env=staging&repoPerEnv=false&gitopsRepo=gitops-staging-infra&deployKeyPath=/deploykey/staging.key;env=production&repoPerEnv=true&gitopsRepo=gitops-production-infra&deployKeyPath=/deploykey/production.key"

	gitopsRepos, err := parseGitopsRepos(validInput)
	if err != nil {
		t.Errorf("Cannot parse gitopsRepos")
	}

	assert.Equal(t, 2, len(gitopsRepos))
	assert.Equal(t, "staging", gitopsRepos[0].env)
	assert.Equal(t, false, gitopsRepos[0].repoPerEnv)
	assert.Equal(t, "gitops-staging-infra", gitopsRepos[0].gitopsRepo)
	assert.Equal(t, "/deploykey/staging.key", gitopsRepos[0].deployKeyPath)
	assert.Equal(t, "production", gitopsRepos[1].env)
	assert.Equal(t, true, gitopsRepos[1].repoPerEnv)
	assert.Equal(t, "gitops-production-infra", gitopsRepos[1].gitopsRepo)
	assert.Equal(t, "/deploykey/production.key", gitopsRepos[1].deployKeyPath)
}
