package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseRepoURL(t *testing.T) {
	host, owner, repo := parseRepoURL("git@github.com:gimlet-io/gimlet-cli.git")
	if host != "github.com" {
		t.Errorf("Must parse host")
	}
	if owner != "gimlet-io" {
		t.Errorf("Must parse owner")
	}
	if repo != "gimlet-cli" {
		t.Errorf("Must parse repo")
	}
}

func Test_generateManifest(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		t.Errorf("Cannot create directory")
	}

	defer os.RemoveAll(dirToWrite)

	env := "staging"
	owner := "gimlet"
	repo := "test-repo"

	gitopsRepoFileName, _, secretFileName, err := generateManifests(
		false,
		env,
		false,
		dirToWrite,
		true,
		true,
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest")
	}

	gitopsRepoFileLocal, err := os.Stat(dirToWrite + fmt.Sprintf("/%s/flux/%s", env, gitopsRepoFileName))
	if err != nil {
		t.Errorf("cannot find gitops repo file in the local directory")
	}

	secretFileLocal, err := os.Stat(dirToWrite + fmt.Sprintf("/%s/flux/%s", env, secretFileName))
	if err != nil {
		t.Errorf("cannot find secret file in the local directory")
	}

	assert.Equal(t, fmt.Sprintf("gitops-repo-%s-%s-%s-%s.yaml", env, owner, repo, env), gitopsRepoFileLocal.Name())
	assert.Equal(t, fmt.Sprintf("deploy-key-%s-%s-%s-%s.yaml", env, owner, repo, env), secretFileLocal.Name())
}
