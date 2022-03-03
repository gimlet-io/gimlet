package gitops

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_parseRepoURL(t *testing.T) {
	host, owner, repo := ParseRepoURL("git@github.com:gimlet-io/gimlet-cli.git")
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

func Test_generateManifestWithoutControllerWithoutSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := false
	env := "staging"
	singleEnv := false
	shouldGenerateKustomizationAndRepo := true
	shouldGenerateDeployKey := true
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	gitopsRepoFileName, _, secretFileName, err := GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate the manifest files, %s", err)
		return
	}

	fluxFile, _ := os.Stat(filepath.Join(dirToWrite, env, "flux", "flux.yaml"))
	if fluxFile != nil {
		t.Errorf("Should not find flux.yaml in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, env, "flux", gitopsRepoFileName))
	if err != nil {
		t.Errorf("Should find gitops-repo file in the local directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, env, "flux", secretFileName))
	if err != nil {
		t.Errorf("Should find deploy-key file in the local directory")
	}
}

func Test_generateManifestWithoutControllerWithSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}
	shouldGenerateController := false
	env := ""
	singleEnv := true
	shouldGenerateKustomizationAndRepo := true
	shouldGenerateDeployKey := true
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	_, _, _, err = GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate the manifest files, %s", err)
		return
	}

	fluxFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "flux.yaml"))
	if fluxFile != nil {
		t.Errorf("Should not find flux.yaml in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "gitops-repo.yaml"))
	if err != nil {
		t.Errorf("Should find gitops-repo.yaml in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "deploy-key.yaml"))
	if err != nil {
		t.Errorf("Should find deploy-key.yaml in the flux directory")
	}
}

func Test_generateManifestWithControllerWithoutSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := true
	env := "staging"
	singleEnv := false
	shouldGenerateKustomizationAndRepo := true
	shouldGenerateDeployKey := true
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	gitopsRepoFileName, _, secretFileName, err := GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	_, err = os.Stat(filepath.Join(dirToWrite, env, "flux", "flux.yaml"))
	if err != nil {
		t.Errorf("Should find flux.yaml in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, env, "flux", gitopsRepoFileName))
	if err != nil {
		t.Errorf("Should find gitops-repo file in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, env, "flux", secretFileName))
	if err != nil {
		t.Errorf("Should find deploy-key file in the flux directory")
	}
}

func Test_generateManifestWithControllerWithSingleEnv(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := true
	env := ""
	singleEnv := true
	shouldGenerateKustomizationAndRepo := true
	shouldGenerateDeployKey := true
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	_, _, _, err = GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "flux.yaml"))
	if err != nil {
		t.Errorf("Should find flux.yaml in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "gitops-repo.yaml"))
	if err != nil {
		t.Errorf("Should find gitopsrepo.yaml file in the flux directory")
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "deploy-key.yaml"))
	if err != nil {
		t.Errorf("Should find deploy-key.yaml in the flux directory")
	}
}

func Test_generateManifestWithoutKustomizationAndRepoWithoutDeployKey(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := false
	env := ""
	singleEnv := true
	shouldGenerateKustomizationAndRepo := false
	shouldGenerateDeployKey := false
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	_, _, _, err = GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	fluxFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "flux.yaml"))
	if fluxFile != nil {
		t.Errorf("Should not find flux.yaml in the flux directory")
	}

	gitopsRepoFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "gitops-repo.yaml"))
	if gitopsRepoFile != nil {
		t.Errorf("Should not find gitops-repo.yaml file in the flux directory")
	}

	secretFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "deploy-key.yaml"))
	if secretFile != nil {
		t.Errorf("Should not find deploy-key.yaml in the flux directory")
	}
}

func Test_generateManifestWithoutKustomizationAndRepoWithDeployKey(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := false
	env := ""
	singleEnv := true
	shouldGenerateKustomizationAndRepo := false
	shouldGenerateDeployKey := true
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	_, _, _, err = GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	fluxFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "flux.yaml"))
	if fluxFile != nil {
		t.Errorf("Should not find flux.yaml in the flux directory")
	}

	gitopsRepoFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "gitops-repo.yaml"))
	if gitopsRepoFile != nil {
		t.Errorf("Should not find gitops-repo.yaml file in the flux directory")
	}

	secretFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "deploy-key.yaml"))
	if secretFile != nil {
		t.Errorf("Should not find deploy-key.yaml in the flux directory")
	}
}

func Test_generateManifestWithKustomizationAndRepoWithoutDeployKey(t *testing.T) {
	dirToWrite, err := ioutil.TempDir("/tmp", "gimlet")
	defer os.RemoveAll(dirToWrite)
	if err != nil {
		t.Errorf("Cannot create directory")
		return
	}

	shouldGenerateController := false
	env := ""
	singleEnv := true
	shouldGenerateKustomizationAndRepo := true
	shouldGenerateDeployKey := false
	gitOpsRepoURL := "git@github.com:gimlet/test-repo.git"

	_, _, _, err = GenerateManifests(
		shouldGenerateController,
		env,
		singleEnv,
		dirToWrite,
		shouldGenerateKustomizationAndRepo,
		shouldGenerateDeployKey,
		gitOpsRepoURL,
		"",
	)
	if err != nil {
		t.Errorf("Cannot generate manifest files, %s", err)
		return
	}

	_, err = os.Stat(filepath.Join(dirToWrite, "flux", "gitops-repo.yaml"))
	if err != nil {
		t.Errorf("Should find gitops-repo.yaml file in the flux directory")
	}

	secretFile, _ := os.Stat(filepath.Join(dirToWrite, "flux", "deploy-key.yaml"))
	if secretFile != nil {
		t.Errorf("Should not find deploy-key.yaml in the flux directory")
	}
}
