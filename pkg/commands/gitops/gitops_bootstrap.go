package gitops

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
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
		&cli.BoolFlag{
			Name:  "no-dependencies",
			Usage: "if you dont't want to use dependencies for Flux",
		},
		&cli.BoolFlag{
			Name:  "kustomization-per-app",
			Usage: "if set, the Kustomization target path will be the Flux folder",
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

	repo, _ := git.PlainOpen(gitopsRepoPath)
	branch, _ := branchName(repo, gitopsRepoPath)

	fmt.Fprintf(os.Stderr, "%v Generating manifests\n", emoji.HourglassNotDone)

	noController := c.Bool("no-controller")
	noDependencies := c.Bool("no-dependencies")
	kustomizationPerApp := c.Bool("kustomization-per-app")
	singleEnv := c.Bool("single-env")
	env := c.String("env")
	gitopsRepoFileName, publicKey, secretFileName, err := gitops.GenerateManifests(
		!noController,
		!noDependencies,
		kustomizationPerApp,
		env,
		singleEnv,
		gitopsRepoPath,
		true,
		true,
		c.String("gitops-repo-url"),
		branch,
	)
	if err != nil {
		return err
	}

	guidingTextFMTPrint := guidingText(gitopsRepoPath, env, publicKey, noController, secretFileName, gitopsRepoFileName)

	fmt.Print(guidingTextFMTPrint)

	return nil
}

func branchName(repo *git.Repository, gitopsRepoPath string) (string, error) {
	if repo == nil {
		return "main", nil
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	if !ref.Name().IsBranch() {
		return "", fmt.Errorf("%s is in a detached state, checkout a branch", gitopsRepoPath)
	}

	return ref.Name().Short(), nil
}

func guidingText(
	gitopsRepoPath string,
	env string,
	publicKey string,
	noController bool,
	secretFileName string,
	gitopsRepoFileName string) string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(
		fmt.Sprintf("%v GitOps configuration written to %s\n\n\n", emoji.CheckMark, filepath.Join(gitopsRepoPath, env, "flux")),
	)
	stringBuilder.WriteString(
		fmt.Sprintf("%v 1) Inspect the configuration files, then commit and push the configuration to git\n", emoji.BackhandIndexPointingRight),
	)
	stringBuilder.WriteString(
		fmt.Sprintf("%v 2) Add the following deploy key to your Git provider\n", emoji.BackhandIndexPointingRight),
	)

	stringBuilder.WriteString(fmt.Sprintf("\n%s\n\n", publicKey))

	stringBuilder.WriteString(
		fmt.Sprintf("%v 3) Apply the gitops manifests on the cluster to start the gitops loop:\n\n", emoji.BackhandIndexPointingRight),
	)

	if !noController {
		stringBuilder.WriteString(fmt.Sprintf("kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", "flux.yaml")))
	}

	stringBuilder.WriteString(
		fmt.Sprintf("kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", secretFileName)),
	)
	stringBuilder.WriteString(
		"kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io\n")
	stringBuilder.WriteString(
		"kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io\n")
	stringBuilder.WriteString(
		fmt.Sprintf("kubectl apply -f %s\n", path.Join(gitopsRepoPath, env, "flux", gitopsRepoFileName)),
	)

	stringBuilder.WriteString(fmt.Sprintf("\n\t Happy Gitopsing%v\n\n", emoji.ConfettiBall))

	return stringBuilder.String()
}
