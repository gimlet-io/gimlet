package nativeGit

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GitopsRepoCache struct {
	cacheRoot               string
	gitopsRepo              string
	gitopsRepos             string
	gitopsRepoDeployKeyPath string
	Repos                   map[string]*git.Repository
	cachePath               string
	stopCh                  chan os.Signal
	waitCh                  chan struct{}
}

func NewGitopsRepoCache(
	cacheRoot string,
	gitopsRepo string,
	gitopsRepos string,
	gitopsRepoDeployKeyPath string,
	stopCh chan os.Signal,
	waitCh chan struct{},
) (*GitopsRepoCache, error) {
	var cachePath string

	parsedGitopsRepos, err := parseGitopsRepos(gitopsRepos)
	if err != nil {
		return nil, err
	}

	repos := map[string]*git.Repository{}
	for _, gitopsRepo := range parsedGitopsRepos {
		repoCachePath, repo, err := CloneToFs(cacheRoot, gitopsRepo.gitopsRepo, gitopsRepo.deployKeyPath)
		if err != nil {
			return nil, err
		}

		repos[gitopsRepo.env] = repo
		cachePath = repoCachePath
	}

	return &GitopsRepoCache{
		cacheRoot:               cacheRoot,
		gitopsRepo:              gitopsRepo,
		gitopsRepos:			 gitopsRepos,
		gitopsRepoDeployKeyPath: gitopsRepoDeployKeyPath,
		Repos:                   repos,
		cachePath:               cachePath,
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
			logrus.Infof("cleaning up git repo cache at %s", r.cachePath)
			TmpFsCleanup(r.cachePath)
			r.waitCh <- struct{}{}
			return
		case <-time.After(30 * time.Second):
		}
	}
}


func (r *GitopsRepoCache) syncGitRepo(repoName string) {
	var publicKeysString string

	parsedGitopsRepos, err := parseGitopsRepos(r.gitopsRepos)
	if err != nil {
		logrus.Errorf("could not parse gitops repositories: %s", err)
		return
	}

	for _, gitopsRepo := range parsedGitopsRepos {
		if gitopsRepo.env == repoName {
			publicKeysString = gitopsRepo.deployKeyPath
		} else {
			publicKeysString = r.gitopsRepoDeployKeyPath
		}
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", publicKeysString, "")
	if err != nil {
		logrus.Errorf("cannot generate public key from private: %s", err.Error())
	}

	w, err := r.Repos[repoName].Worktree()
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
	return r.Repos[repoName]
}

func (r *GitopsRepoCache) InstanceForWrite(repoName string) (*git.Repository, string, string, error) {
	var tmpDirName, deployKeyPath string
	var err error

	parsedGitopsRepos, err := parseGitopsRepos(r.gitopsRepos)
	if err != nil {
		errors.WithMessage(err, "couldn't parse gitops repositories")
	}

	for _, repo := range parsedGitopsRepos {
		if repo.env == repoName {
			tmpDirName = repoName
			deployKeyPath = repo.deployKeyPath
		} else {
			tmpDirName = r.cacheRoot
			deployKeyPath = r.gitopsRepoDeployKeyPath
		}
	}

	tmpPath, err := ioutil.TempDir(tmpDirName, "gitops-cow-")
	if err != nil {
		errors.WithMessage(err, "couldn't get temporary directory")
	}

	err = copy.Copy(r.cachePath, tmpPath)
	if err != nil {
		errors.WithMessage(err, "could not make copy of repo")
	}

	copiedRepo, err := git.PlainOpen(tmpPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("cannot open git repository at %s: %s", tmpPath, err)
	}

	return copiedRepo, tmpPath, deployKeyPath, nil
}

func (r *GitopsRepoCache) CleanupWrittenRepo(path string) error {
	return os.RemoveAll(path)
}

func (r *GitopsRepoCache) Invalidate(repoName string) {
	r.syncGitRepo(repoName)
}

type gitopsRepoConfig struct {
	env           string
	repoPerEnv    bool
	gitopsRepo    string
	deployKeyPath string
}

func parseGitopsRepos(gitopsReposString string) ([]*gitopsRepoConfig, error) {
	var gitopsRepos []*gitopsRepoConfig
	splitGitopsRepos := strings.Split(gitopsReposString, ";")

	for _, gitopsReposString := range splitGitopsRepos {
		if gitopsReposString == "" {
			continue
		}
		parsedGitopsReposString, err := url.ParseQuery(gitopsReposString)
		if err != nil {
			return nil, fmt.Errorf("invalid gitopsRepos format: %s", err)
		}
		repoPerEnv, err := strconv.ParseBool(parsedGitopsReposString.Get("repoPerEnv"))
		if err != nil {
			return nil, fmt.Errorf("invalid gitopsRepos format: %s", err)
		}

		singleGitopsRepo := &gitopsRepoConfig{
			env:           parsedGitopsReposString.Get("env"),
			repoPerEnv:    repoPerEnv,
			gitopsRepo:    parsedGitopsReposString.Get("gitopsRepo"),
			deployKeyPath: parsedGitopsReposString.Get("deployKeyPath"),
		}
		gitopsRepos = append(gitopsRepos, singleGitopsRepo)
	}

	return gitopsRepos, nil
}
