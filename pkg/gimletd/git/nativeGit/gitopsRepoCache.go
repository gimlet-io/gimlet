package nativeGit

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GitopsRepoCache struct {
	cacheRoot               string
	parsedGitopsRepos       []*config.GitopsRepoConfig
	gitopsRepoDeployKeyPath string
	defaultRepo				*git.Repository
	Repos                   map[string]*git.Repository
	defaultCachePath		string
	cachePaths              map[string]string
	stopCh                  chan os.Signal
	waitCh                  chan struct{}
}

func NewGitopsRepoCache(
	cacheRoot string,
	gitopsRepo string,
	parsedGitopsRepos []*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	stopCh chan os.Signal,
	waitCh chan struct{},
) (*GitopsRepoCache, error) {
	defaultCachePath, defaultRepo, err := CloneToFs(cacheRoot, gitopsRepo, gitopsRepoDeployKeyPath)
	if err != nil {
		return nil, err
	}

	cachePaths := map[string]string{}
	repos := map[string]*git.Repository{}
	for _, gitopsRepo := range parsedGitopsRepos {
		repoCachePath, repo, err := CloneToFs(cacheRoot, gitopsRepo.GitopsRepo, gitopsRepo.DeployKeyPath)
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
		defaultRepo:			 defaultRepo,
		Repos:                   repos,
		defaultCachePath:		 defaultCachePath,
		cachePaths:              cachePaths,
		stopCh:                  stopCh,
		waitCh:                  waitCh,
	}, nil
}

func (r *GitopsRepoCache) Run() {
	for {
		for repoName := range r.Repos {
			r.syncGitRepo(repoName)
		}

		select {
		case <-r.stopCh:
			logrus.Infof("cleaning up git repo cache at %s", r.defaultCachePath)
			TmpFsCleanup(r.defaultCachePath)
			for _, cachePath := range r.cachePaths {
				logrus.Infof("cleaning up git repo cache at %s", cachePath)
				TmpFsCleanup(cachePath)
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

	w, err := r.repoToUse(repoName).Worktree()
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
	return r.repoToUse(repoName)
}

func (r *GitopsRepoCache) InstanceForWrite(repoName string) (*git.Repository, string, string, error) {
	tmpPath, err := ioutil.TempDir(r.cacheRoot, "gitops-cow-")
	if err != nil {
		errors.WithMessage(err, "couldn't get temporary directory")
	}
 
	cachePath := r.defaultCachePath
	for cachePathName, cachePathContent := range r.cachePaths {
		if cachePathName == repoName {
			cachePath = cachePathContent
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

func (r *GitopsRepoCache) repoToUse(repoName string) *git.Repository {
	repoToRead := r.defaultRepo
	for repoistoryName, repository := range r.Repos {
		if repoistoryName == repoName {
			repoToRead = repository
		}
	}
	
	return repoToRead
}
