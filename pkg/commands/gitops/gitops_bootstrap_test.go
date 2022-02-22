package gitops

import (
	"io/ioutil"
	"os"
	"testing"
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

	gitopsRepoFileName, _, secretFileName, err := generateManifests(
		false,
		"staging",
		false,
		dirToWrite,
		true,
		true,
		"git@github.com:dzsak/gitops-apps.git",
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest")
	}

	fileName, err := os.Stat(dirToWrite + "/staging/flux/gitops-repo-staging.yaml")
	if err != nil {
		t.Errorf("cannot find")
	}

	secretFile, err := os.Stat(dirToWrite + "/staging/flux/deploy-key-staging.yaml")
	if err != nil {
		t.Errorf("cannot find")
	}

	assertEqual(t, secretFileName, secretFile.Name())
	assertEqual(t, gitopsRepoFileName, fileName.Name())
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("%s != %s", a, b)
	}
}
