package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
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

func Test_generateManifestWithoutControllerWithoutSingleEnv(t *testing.T) {
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

func Test_generateManifestWithSingleEnvWithoutController(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		t.Errorf("Cannot create directory")
	}

	defer os.RemoveAll(dirToWrite)

	owner := "gimlet"
	repo := "test-repo"

	gitopsRepoFileName, _, secretFileName, err := generateManifests(
		false,
		"",
		true,
		dirToWrite,
		true,
		true,
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest")
	}

	assert.Equal(t, "gitops-repo.yaml", gitopsRepoFileName)
	assert.Equal(t, "deploy-key.yaml", secretFileName)

	gitopsRepoFileLocal, err := os.Stat(dirToWrite + "/flux/gitops-repo.yaml")
	if err != nil {
		t.Errorf("cannot find gitops repo file in the local directory")
	}

	secretFileLocal, err := os.Stat(dirToWrite + "/flux/deploy-key.yaml")
	if err != nil {
		t.Errorf("cannot find secret file in the local directory")
	}

	assert.Equal(t, "gitops-repo.yaml", gitopsRepoFileLocal.Name())
	assert.Equal(t, "deploy-key.yaml", secretFileLocal.Name())
}

func Test_generateManifestWithControllerWithoutSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		t.Errorf("Cannot create directory")
	}

	defer os.RemoveAll(dirToWrite)

	env := "staging"
	owner := "gimlet"
	repo := "test-repo"

	_, _, _, err = generateManifests(
		true,
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

	fluxFileLocal, err := os.Stat(dirToWrite + fmt.Sprintf("/%s/flux/flux.yaml", env))
	if err != nil {
		t.Errorf("cannot find flux file in the local directory")
	}

	assert.Equal(t, "flux.yaml", fluxFileLocal.Name())
}

func Test_generateManifestWithControllerWithSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		t.Errorf("Cannot create directory")
	}

	defer os.RemoveAll(dirToWrite)

	owner := "gimlet"
	repo := "test-repo"

	_, _, _, err = generateManifests(
		true,
		"",
		true,
		dirToWrite,
		true,
		true,
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest")
	}

	fluxFileLocal, err := os.Stat(dirToWrite + "/flux/flux.yaml")
	if err != nil {
		t.Errorf("cannot find flux file in the local directory")
	}
	assert.Equal(t, "flux.yaml", fluxFileLocal.Name())

	_, err = os.Stat(dirToWrite + "/flux/gitops-repo.yaml")
	if err != nil {
		t.Errorf("cannot find gitops repo file in the local directory")
	}
	_, err = os.Stat(dirToWrite + "/flux/deploy-key.yaml")
	if err != nil {
		t.Errorf("cannot find secret file in the local directory")
	}
}

func Test_guidingText(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		t.Errorf("Cannot create directory")
	}

	defer os.RemoveAll(dirToWrite)

	env := "staging"
	owner := "gimlet"
	repo := "test-repo"
	gitopsRepoPathName := "gitops-repo-path"
	publicKey := "12345"

	gitopsRepoFileName, _, secretFileName, _ := generateManifests(
		false,
		env,
		false,
		dirToWrite,
		true,
		true,
		fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		"",
	)

	stderrString := capturer.CaptureOutput(func() {
		(GuidingText(gitopsRepoPathName, env, publicKey, false, secretFileName, gitopsRepoFileName))
	})

	secretFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/" + env + "/flux/" + secretFileName
	gitopsRepoFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/" + env + "/flux/" + gitopsRepoFileName

	if !strings.Contains(stderrString, secretFileNameGuidingText) {
		t.Errorf("Stderr does not contain specified string in deploy key path")
	}

	if !strings.Contains(stderrString, gitopsRepoFileNameGuidingText) {
		t.Errorf("Stderr does not contain specified string in gitops repo path")
	}
}
