package gitops

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops/sync"
	"github.com/stretchr/testify/assert"
)

func Test_guidingTextWithoutController(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	gitopsRepoPathName := "gitops-repo-path"
	publicKey := "12345"
	noController := true

	opts := gitops.ManifestOpts{
		ShouldGenerateController:           !noController,
		ShouldGenerateDependencies:         true,
		KustomizationPerApp:                true,
		Env:                                "staging",
		SingleEnv:                          false,
		GitopsRepoPath:                     dirToWrite,
		ShouldGenerateKustomizationAndRepo: true,
		ShouldGenerateDeployKey:            true,
		GitopsRepoUrl:                      "git@github.com:gimlet/test-repo.git",
		Branch:                             "",
	}

	gitopsRepoFileName, _, secretFileName, err := gitops.GenerateManifests(opts)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	guidingTextString := guidingText(gitopsRepoPathName, opts.Env, publicKey, noController, secretFileName, gitopsRepoFileName)

	secretFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/" + opts.Env + "/flux/" + secretFileName
	gitopsRepoFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/" + opts.Env + "/flux/" + gitopsRepoFileName
	withoutControllerGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/" + opts.Env + "/flux/flux.yaml"

	if !strings.Contains(guidingTextString, secretFileNameGuidingText) {
		t.Errorf("Should contain specified string in deploy key path")
	}

	if !strings.Contains(guidingTextString, gitopsRepoFileNameGuidingText) {
		t.Errorf("Should contain specified string in gitops repo path")
	}

	if strings.Contains(guidingTextString, withoutControllerGuidingText) {
		t.Errorf("Should not contain line about flux.yaml creation")
	}
}

func Test_guidingTextWithoutControllerAndSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	gitopsRepoPathName := "gitops-repo-path"
	publicKey := "12345"
	noController := true

	opts := gitops.ManifestOpts{
		ShouldGenerateController:           !noController,
		ShouldGenerateDependencies:         true,
		KustomizationPerApp:                true,
		Env:                                "",
		SingleEnv:                          true,
		GitopsRepoPath:                     dirToWrite,
		ShouldGenerateKustomizationAndRepo: true,
		ShouldGenerateDeployKey:            true,
		GitopsRepoUrl:                      "git@github.com:gimlet/test-repo.git",
		Branch:                             "",
	}

	gitopsRepoFileName, _, secretFileName, err := gitops.GenerateManifests(opts)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	guidingTextString := guidingText(gitopsRepoPathName, opts.Env, publicKey, noController, secretFileName, gitopsRepoFileName)

	secretFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/flux/" + secretFileName
	gitopsRepoFileNameGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/flux/" + gitopsRepoFileName
	withoutControllerGuidingText := "kubectl apply -f " + gitopsRepoPathName + "/flux/flux.yaml"

	if !strings.Contains(guidingTextString, secretFileNameGuidingText) {
		t.Errorf("Should contain specified string in deploy key path")
	}

	if !strings.Contains(guidingTextString, gitopsRepoFileNameGuidingText) {
		t.Errorf("Should contain specified string in gitops repo path")
	}

	if strings.Contains(guidingTextString, withoutControllerGuidingText) {
		t.Errorf("Should not contain line about flux.yaml creation")
	}
}

func Test_guidingTextWithController(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	gitopsRepoPathName := "gitops-repo-path"
	publicKey := "12345"
	noController := false

	opts := gitops.ManifestOpts{
		ShouldGenerateController:           !noController,
		ShouldGenerateDependencies:         true,
		KustomizationPerApp:                true,
		Env:                                "staging",
		SingleEnv:                          false,
		GitopsRepoPath:                     dirToWrite,
		ShouldGenerateKustomizationAndRepo: true,
		ShouldGenerateDeployKey:            true,
		GitopsRepoUrl:                      "git@github.com:gimlet/test-repo.git",
		Branch:                             "",
	}

	gitopsRepoFileName, _, secretFileName, err := gitops.GenerateManifests(opts)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	guidingTextString := guidingText(gitopsRepoPathName, opts.Env, publicKey, noController, secretFileName, gitopsRepoFileName)

	guidingTextWithoutControllerText := "kubectl apply -f " + gitopsRepoPathName + "/" + opts.Env + "/flux/flux.yaml"

	if !strings.Contains(guidingTextString, guidingTextWithoutControllerText) {
		t.Errorf("Should contain line about flux.yaml creation")
	}
}

func Test_DependenciesPathWithTargetPath(t *testing.T) {
	targetPath := "staging"
	dependenciesPath := sync.DependenciesPath(targetPath)

	assert.Equal(t, dependenciesPath, "./staging/dependencies", "The dependencies path should be './staging/dependencies'")
}

func Test_DependenciesPathWithoutTargetPath(t *testing.T) {
	targetPath := ""
	dependenciesPath := sync.DependenciesPath(targetPath)

	assert.Equal(t, dependenciesPath, "./dependencies", "The path should be './dependencies'")
}
