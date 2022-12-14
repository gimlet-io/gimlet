package gitops

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GitopsRepoCache struct {
	cacheRoot               string
	parsedGitopsRepos       map[string]*config.GitopsRepoConfig
	gitopsRepoDeployKeyPath string
	defaultRepo             *git.Repository
	defaultRepoName         string
	repos                   map[string]*git.Repository
	defaultCachePath        string
	cachePaths              map[string]string
	stopCh                  chan os.Signal
	waitCh                  chan struct{}
}

func NewGitopsRepoCache(
	cacheRoot string,
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	gitSSHAddressFormat string,
	stopCh chan os.Signal,
	waitCh chan struct{},

) (*GitopsRepoCache, error) {
	var defaultRepo *git.Repository
	var defaultCachePath string
	var err error
	if gitopsRepo != "" && gitopsRepoDeployKeyPath != "" {
		defaultCachePath, defaultRepo, err = nativeGit.CloneToFs(
			cacheRoot,
			gitopsRepo,
			gitopsRepoDeployKeyPath,
			gitSSHAddressFormat,
		)
		if err != nil {
			return nil, err
		}
	}

	cachePaths := map[string]string{}
	repos := map[string]*git.Repository{}
	for _, gitopsRepo := range parsedGitopsRepos {
		repoCachePath, repo, err := nativeGit.CloneToFs(
			cacheRoot,
			gitopsRepo.GitopsRepo,
			gitopsRepo.DeployKeyPath,
			gitSSHAddressFormat,
		)
		if err != nil {
			return nil, err
		}

		repos[gitopsRepo.GitopsRepo] = repo
		cachePaths[gitopsRepo.GitopsRepo] = repoCachePath
	}

	return &GitopsRepoCache{
		cacheRoot:               cacheRoot,
		parsedGitopsRepos:       parsedGitopsRepos,
		gitopsRepoDeployKeyPath: gitopsRepoDeployKeyPath,
		defaultRepo:             defaultRepo,
		defaultRepoName:         gitopsRepo,
		repos:                   repos,
		defaultCachePath:        defaultCachePath,
		cachePaths:              cachePaths,
		stopCh:                  stopCh,
		waitCh:                  waitCh,
	}, nil
}

func (r *GitopsRepoCache) Run() {
	for {
		if r.defaultRepoName != "" {
			r.syncGitRepo(r.defaultRepoName)
		}
		for repoName := range r.repos {
			r.syncGitRepo(repoName)
		}

		select {
		case <-r.stopCh:
			if r.defaultCachePath != "" {
				logrus.Infof("cleaning up git repo cache at %s", r.defaultCachePath)
				nativeGit.TmpFsCleanup(r.defaultCachePath)
			}
			for _, cachePath := range r.cachePaths {
				logrus.Infof("cleaning up git repo cache at %s", cachePath)
				nativeGit.TmpFsCleanup(cachePath)
			}
			r.waitCh <- struct{}{}
			return
		case <-time.After(30 * time.Second):
		}
	}
}

func (r *GitopsRepoCache) syncGitRepo(repoName string) {
	publicKeysString := r.gitopsRepoDeployKeyPath
	for _, gitopsRepo := range r.parsedGitopsRepos {
		if gitopsRepo.GitopsRepo == repoName {
			publicKeysString = gitopsRepo.DeployKeyPath
		}
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", publicKeysString, "")
	if err != nil {
		logrus.Errorf("cannot generate public key from private: %s", err.Error())
	}

	var repo *git.Repository
	if repoInMap, exists := r.repos[repoName]; exists {
		repo = repoInMap
	} else {
		repo = r.defaultRepo
	}

	w, err := repo.Worktree()
	if err != nil {
		logrus.Errorf("could not get worktree: %s", err)
		return
	}

	w.Pull(&git.PullOptions{
		Auth:       publicKeys,
		RemoteName: "origin",
	})
	if err == git.NoErrAlreadyUpToDate {
		return
	}
	if err != nil {
		logrus.Errorf("could not fetch: %s", err)
	}
}

func (r *GitopsRepoCache) InstanceForRead(repoName string) *git.Repository {
	if repoInMap, exists := r.repos[repoName]; exists {
		return repoInMap
	}

	return r.defaultRepo
}

func (r *GitopsRepoCache) InstanceForWrite(repoName string) (*git.Repository, string, string, error) {
	tmpPath, err := ioutil.TempDir(r.cacheRoot, "gitops-cow-")
	if err != nil {
		errors.WithMessage(err, "couldn't get temporary directory")
	}

	cachePath := r.defaultCachePath
	for gitopsRepo, gitopsRepoCachePath := range r.cachePaths {
		if gitopsRepo == repoName {
			cachePath = gitopsRepoCachePath
		}
	}

	err = copy.Copy(cachePath, tmpPath)
	if err != nil {
		errors.WithMessage(err, "could not make copy of repo")
	}

	copiedRepo, err := git.PlainOpen(tmpPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("cannot open git repository at %s: %s", tmpPath, err)
	}

	deployKeyPath := r.gitopsRepoDeployKeyPath
	for _, repo := range r.parsedGitopsRepos {
		if repo.GitopsRepo == repoName {
			deployKeyPath = repo.DeployKeyPath
		}
	}

	return copiedRepo, tmpPath, deployKeyPath, nil
}

func (r *GitopsRepoCache) CleanupWrittenRepo(path string) error {
	return os.RemoveAll(path)
}

func (r *GitopsRepoCache) Invalidate(repoName string) {
	r.syncGitRepo(repoName)
}
