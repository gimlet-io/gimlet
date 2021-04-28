package gitops

import (
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/fluxcd/flux2/pkg/manifestgen/sync"
	"github.com/fluxcd/pkg/ssh"
	"github.com/gimlet-io/gimletd/githelper"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

var gitopsBootstrapCmd = cli.Command{
	Name:  "bootstrap",
	Usage: "Bootstraps the gitops controller for an environment",
	UsageText: `gimlet gitops bootstrap \
     --env staging \
     --gitops-repo-url git@github.com:<user>/<repo>.git`,
	Action: bootstrap,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to bootstrap (mandatory)",
			Required: true,
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
	},
}

func bootstrap(c *cli.Context) error {
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

	empty, err := githelper.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("there are changes in the gitops repo. Commit them first then try again")
	}

	env := c.String("env")
	gitopsRepoUrl := c.String("gitops-repo-url")
	host, owner, repoName := parseRepoURL(gitopsRepoUrl)

	installOpts := install.MakeDefaultOptions()
	installOpts.ManifestFile = "flux.yaml"
	installOpts.TargetPath = env

	fmt.Fprintf(os.Stderr, "%v Generating manifests\n", emoji.HourglassNotDone)

	installManifest, err := install.Generate(installOpts, "")
	if err != nil {
		return fmt.Errorf("cannot generate installation manifests %s", err)
	}
	installManifest.Path = path.Join(env, "flux", installOpts.ManifestFile)
	_, err = installManifest.WriteFile(gitopsRepoPath)
	if err != nil {
		return fmt.Errorf("cannot write installation manifests %s", err)
	}

	syncOpts := sync.Options{
		Interval:     15 * time.Second,
		URL:          fmt.Sprintf("ssh://git@%s/%s/%s", host, owner, repoName),
		Name:         "gitops-repo",
		Namespace:    "flux-system",
		Branch:       "main",
		ManifestFile: "gitops-repo.yaml",
	}

	syncOpts.TargetPath = env
	syncManifest, err := sync.Generate(syncOpts)
	if err != nil {
		return fmt.Errorf("cannot generate git manifests %s", err)
	}
	syncManifest.Path = path.Join(env, "flux", syncOpts.ManifestFile)
	_, err = syncManifest.WriteFile(gitopsRepoPath)
	if err != nil {
		return fmt.Errorf("cannot write git manifests %s", err)
	}

	fmt.Fprintf(os.Stderr, "%v Generating deploy key\n", emoji.HourglassNotDone)

	publicKey, deployKeySecret, err := generateDeployKey(host)
	if err != nil {
		return fmt.Errorf("cannot generate deploy key %s", err)
	}
	err = ioutil.WriteFile(path.Join(gitopsRepoPath, env, "flux", "deploy-key.yaml"), deployKeySecret, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write deploy key %s", err)
	}

	err = githelper.StageFolder(repo, filepath.Join(env, "flux"))
	if err != nil {
		return err
	}

	empty, err = githelper.NothingToCommit(repo)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gitMessage := fmt.Sprintf("[Gimlet CLI bootstrap] %s", env)
	_, err = githelper.Commit(repo, gitMessage)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v GitOps configuration written to %s\n\n\n", emoji.CheckMark, filepath.Join(gitopsRepoPath, env, "flux"))

	fmt.Fprintf(os.Stderr, "%v 1) Push the configuration to git\n", emoji.BackhandIndexPointingRight)
	fmt.Fprintf(os.Stderr, "%v 2) Add the following deploy key to your Git provider\n", emoji.BackhandIndexPointingRight)

	fmt.Printf("\n%s\n", publicKey)

	fmt.Fprintf(os.Stderr, "%v 3) Apply the gitops manifests on the cluster to start the gitops loop:\n\n", emoji.BackhandIndexPointingRight)

	fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", "flux.yaml"))
	fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", "deploy-key.yaml"))
	fmt.Fprintf(os.Stderr, "kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io\n")
	fmt.Fprintf(os.Stderr, "kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io\n")
	fmt.Fprintf(os.Stderr, "kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", "gitops-repo.yaml"))

	fmt.Fprintf(os.Stderr, "\n\t Happy Gitopsing%v\n\n", emoji.ConfettiBall)

	return nil
}

func generateDeployKey(host string) (string, []byte, error) {
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
			Name:      "gitops-repo",
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
