package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/fluxcd/pkg/ssh"
	"github.com/gimlet-io/gimlet-cli/commands/gitops/sync"
	"github.com/gimlet-io/gimletd/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var gitopsBootstrapCmd = cli.Command{
	Name:  "bootstrap",
	Usage: "Bootstraps the gitops controller for an environment",
	UsageText: `gimlet gitops bootstrap \
     --env staging \
     --gitops-repo-url git@github.com:<user>/<repo>.git`,
	Action: Bootstrap,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "env",
			Usage: "environment to bootstrap",
		},
		&cli.BoolFlag{
			Name:  "single-env",
			Usage: "if the repo holds manifests from a single environment",
		},
		&cli.StringFlag{
			Name:     "gitops-repo-url",
			Usage:    "URL of the gitops repo (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gitops-repo-path",
			Usage: "path to the working copy of the gitops repo, default: current dir",
		},
		&cli.BoolFlag{
			Name:  "no-controller",
			Usage: "to not bootstrap the FluxV2 gitops controller, only the GitRepository and Kustomization to add a new source",
		},
	},
}

func Bootstrap(c *cli.Context) error {
	gitopsRepoPath := c.String("gitops-repo-path")
	if gitopsRepoPath == "" {
		gitopsRepoPath, _ = os.Getwd()
	}
	gitopsRepoPath, err := filepath.Abs(gitopsRepoPath)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(gitopsRepoPath)
	if err == git.ErrRepositoryNotExists {
		return fmt.Errorf("%s is not a git repo\n", gitopsRepoPath)
	}
	branch, _ := branchName(err, repo, gitopsRepoPath)
	if branch == "" {
		_, err = nativeGit.Commit(repo, "Initial commit")
		if err != nil {
			return err
		}
		branch, _ = branchName(err, repo, gitopsRepoPath)
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are changes in the gitops repo. Commit them first then try again")
	}

	fmt.Fprintf(os.Stderr, "%v Generating manifests\n", emoji.HourglassNotDone)

	noController := c.Bool("no-controller")
	singleEnv := c.Bool("single-env")
	env := c.String("env")
	gitopsRepositoryName, publicKey, secretFileName, err := generateManifests(
		noController,
		env,
		singleEnv,
		gitopsRepoPath,
		true,
		c.String("gitops-repo-url"),
		branch,
	)
	if err != nil {
		return err
	}

	err = nativeGit.StageFolder(repo, filepath.Join(env, "flux"))
	if err != nil {
		return err
	}

	empty, err = nativeGit.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gitMessage := fmt.Sprintf("[Gimlet CLI] Bootstrapping %s", env)
	if singleEnv {
		gitMessage = "[Gimlet CLI] Bootstrapping"
	}
	_, err = nativeGit.Commit(repo, gitMessage)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v GitOps configuration written to %s\n\n\n", emoji.CheckMark, filepath.Join(gitopsRepoPath, env, "flux"))

	fmt.Fprintf(os.Stderr, "%v 1) Push the configuration to git\n", emoji.BackhandIndexPointingRight)
	fmt.Fprintf(os.Stderr, "%v 2) Add the following deploy key to your Git provider\n", emoji.BackhandIndexPointingRight)

	fmt.Printf("\n%s\n", publicKey)

	fmt.Fprintf(os.Stderr, "%v 3) Apply the gitops manifests on the cluster to start the gitops loop:\n\n", emoji.BackhandIndexPointingRight)

	if !noController {
		fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", "flux.yaml"))
	}
	fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", secretFileName))
	fmt.Fprintf(os.Stderr, "kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io\n")
	fmt.Fprintf(os.Stderr, "kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io\n")
	fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", gitopsRepositoryName+".yaml"))

	fmt.Fprintf(os.Stderr, "\n\t Happy Gitopsing%v\n\n", emoji.ConfettiBall)

	return nil
}

func generateManifests(
	noController bool,
	env string,
	singleEnv bool,
	gitopsRepoPath string,
	generateKustomizationDeployKeyAndRepo bool,
	gitopsRepoUrl string,
	branch string,
) (string, string, string, error) {
	publicKey := ""
	gitopsRepositoryName := ""
	secretFileName := ""

	installOpts := install.MakeDefaultOptions()
	installOpts.ManifestFile = "flux.yaml"
	installOpts.TargetPath = env

	if !singleEnv && env == "" {
		return "", "", "", fmt.Errorf("either `--env` or `--single-env` is mandatory")
	}
	if singleEnv && env != "" {
		return "", "", "", fmt.Errorf("`--env` and `--single-env` are mutually exclusive")
	}

	if singleEnv {
		env = "."
	}

	if !noController {
		installManifest, err := install.Generate(installOpts, "")
		if err != nil {
			return "", "", "", fmt.Errorf("cannot generate installation manifests %s", err)
		}
		installManifest.Path = path.Join(env, "flux", installOpts.ManifestFile)
		_, err = installManifest.WriteFile(gitopsRepoPath)
		if err != nil {
			return "", "", "", fmt.Errorf("cannot write installation manifests %s", err)
		}
	}

	if generateKustomizationDeployKeyAndRepo {
		host, owner, repoName := parseRepoURL(gitopsRepoUrl)
		gitopsRepositoryName = fmt.Sprintf("gitops-repo-%s", strings.ToLower(env))
		if singleEnv {
			gitopsRepositoryName = "gitops-repo"
		}
		syncOpts := sync.Options{
			Interval:     15 * time.Second,
			URL:          fmt.Sprintf("ssh://git@%s/%s/%s", host, owner, repoName),
			Name:         gitopsRepositoryName,
			Secret:       gitopsRepositoryName,
			Namespace:    "flux-system",
			Branch:       branch,
			ManifestFile: gitopsRepositoryName + ".yaml",
		}

		syncOpts.TargetPath = env
		if singleEnv {
			syncOpts.TargetPath = ""
		}
		syncManifest, err := sync.Generate(syncOpts)
		if err != nil {
			return "", "", "", fmt.Errorf("cannot generate git manifests %s", err)
		}
		syncManifest.Path = path.Join(env, "flux", syncOpts.ManifestFile)
		_, err = syncManifest.WriteFile(gitopsRepoPath)
		if err != nil {
			return "", "", "", fmt.Errorf("cannot write git manifests %s", err)
		}

		fmt.Fprintf(os.Stderr, "%v Generating deploy key\n", emoji.HourglassNotDone)
		secretFileName = fmt.Sprintf("deploy-key-%s.yaml", env)
		if singleEnv {
			secretFileName = "deploy-key.yaml"
		}

		pKey, deployKeySecret, err := generateDeployKey(host, gitopsRepositoryName)
		publicKey = pKey
		if err != nil {
			return "", "", "", fmt.Errorf("cannot generate deploy key %s", err)
		}
		err = ioutil.WriteFile(path.Join(gitopsRepoPath, env, "flux", secretFileName), deployKeySecret, os.ModePerm)
		if err != nil {
			return "", "", "", fmt.Errorf("cannot write deploy key %s", err)
		}
	}

	return gitopsRepositoryName, publicKey, secretFileName, nil
}

func branchName(err error, repo *git.Repository, gitopsRepoPath string) (string, error) {
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	if !ref.Name().IsBranch() {
		return "", fmt.Errorf("%s is in a detached state, checkout a branch\n", gitopsRepoPath)
	}

	return ref.Name().Short(), nil
}

func generateDeployKey(host string, name string) (string, []byte, error) {
	privateKeyBytes, publicKeyBytes := generateKeyPair()

	hostKey, err := ssh.ScanHostKey(host+":22", 30*time.Second)
	if err != nil {
		return "", []byte(""), err
	}

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "flux-system",
		},
		StringData: map[string]string{
			"identity":     string(privateKeyBytes),
			"identity.pub": string(publicKeyBytes),
			"known_hosts":  string(hostKey),
		},
	}

	yamlString, err := yaml.Marshal(secret)
	return string(publicKeyBytes), yamlString, err
}

func parseRepoURL(url string) (string, string, string) {
	host := strings.Split(url, ":")[0]
	host = strings.Split(host, "@")[1]

	owner := strings.Split(url, ":")[1]
	owner = strings.Split(owner, "/")[0]

	repo := strings.Split(url, ":")[1]
	repo = strings.Split(repo, "/")[1]
	repo = strings.TrimSuffix(repo, ".git")

	return host, owner, repo
}
