package nativeGit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dashboardConfig "github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var FetchRefSpec = []config.RefSpec{
	"refs/heads/*:refs/remotes/origin/*",
}

type RepoCache struct {
	tokenManager *customScm.NonImpersonatedTokenManager
	repos        map[string]repoData
	stopCh       chan struct{}
	invalidateCh chan string
	cachePath    string

	// For webhook registration
	goScmHelper *genericScm.GoScmHelper
	config      *dashboardConfig.PersistentConfig

	clientHub *streaming.ClientHub

	lock sync.Mutex
}

type repoData struct {
	repo        *git.Repository
	withHistory bool
}

func NewRepoCache(
	tokenManager *customScm.NonImpersonatedTokenManager,
	stopCh chan struct{},
	cachePath string,
	goScmHelper *genericScm.GoScmHelper,
	config *dashboardConfig.PersistentConfig,
	clientHub *streaming.ClientHub,
) (*RepoCache, error) {
	repoCache := &RepoCache{
		tokenManager: tokenManager,
		repos:        map[string]repoData{},
		stopCh:       stopCh,
		invalidateCh: make(chan string),
		cachePath:    cachePath,
		goScmHelper:  goScmHelper,
		config:       config,
		clientHub:    clientHub,
	}

	const DirRwxRxR = 0754
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		os.MkdirAll(cachePath, DirRwxRxR)
	}
	paths, err := os.ReadDir(cachePath)
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}

	for _, fileInfo := range paths {
		if !fileInfo.IsDir() {
			continue
		}

		path := filepath.Join(cachePath, fileInfo.Name())
		repo, err := git.PlainOpen(path)
		if err != nil {
			logrus.Warnf("cannot open git repository at %s: %s", path, err)
			continue
		}

		repoCache.repos[strings.ReplaceAll(fileInfo.Name(), "%", "/")] = repoData{repo, false}
	}

	return repoCache, nil
}

func (r *RepoCache) Run() {
	for {
		for repoName, _ := range r.repos {
			r.syncGitRepo(repoName)
		}

		select {
		case <-r.stopCh:
			logrus.Info("stopping")
			return
		case repoName := <-r.invalidateCh:
			logrus.Infof("received cache invalidate message for %s", repoName)
			r.syncGitRepo(repoName)
		case <-time.After(30 * time.Second):
		}
	}
}

func (r *RepoCache) syncGitRepo(repoName string) {
	tokenManager := *r.tokenManager
	token, user, err := tokenManager.Token()
	if err != nil {
		logrus.Errorf("couldn't get scm token: %s", err)
		return
	}

	if _, ok := r.repos[repoName]; !ok {
		logrus.Warnf("could not get repo by name from cache: %s", repoName)
		return // preventing a race condition in cleanup
	}

	repoData := r.repos[repoName]
	repo := repoData.repo

	opts := &git.FetchOptions{
		RefSpecs: FetchRefSpec,
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 100,
		Tags:  git.NoTags,
		Prune: true,
	}
	if repoData.withHistory {
		opts.Depth = 0
	}

	err = repo.Fetch(opts)
	if err == git.NoErrAlreadyUpToDate {
		return
	}
	if err != nil {
		logrus.Errorf("could not fetch: %s", err)
		r.cleanRepo(repoName)
	}

	w, err := repo.Worktree()
	if err != nil {
		logrus.Errorf("could not get working copy: %s", err)
		r.cleanRepo(repoName)
		return
	}

	headBranch, err := HeadBranch(repo)
	if err != nil {
		logrus.Errorf("cannot get head branch: %s", err)
		r.cleanRepo(repoName)
		return
	}

	branchHeadHash := BranchHeadHash(repo, headBranch)
	err = w.Reset(&git.ResetOptions{
		Commit: branchHeadHash,
		Mode:   git.HardReset,
	})
	if err != nil {
		logrus.Errorf("could not reset: %s", err)
		r.cleanRepo(repoName)
		return
	}

	if r.clientHub == nil {
		return
	}
	jsonString, _ := json.Marshal(streaming.StaleRepoDataEvent{
		Repo:           repoName,
		StreamingEvent: streaming.StreamingEvent{Event: streaming.StaleRepoDataEventString},
	})
	r.clientHub.Broadcast <- jsonString
}

func (r *RepoCache) cleanRepo(repoName string) {
	r.lock.Lock()
	delete(r.repos, repoName)
	r.lock.Unlock()
}

func (r *RepoCache) InstanceForRead(repoName string) (instance *git.Repository, err error) {
	return r.instanceForRead(repoName, false)
}

func (r *RepoCache) InstanceForReadWithHistory(repoName string) (instance *git.Repository, err error) {
	return r.instanceForRead(repoName, true)
}

func (r *RepoCache) instanceForRead(repoName string, withHistory bool) (instance *git.Repository, err error) {
	r.lock.Lock()
	if repoData, ok := r.repos[repoName]; ok {
		if withHistory && !repoData.withHistory {
			repoData, err = r.clone(repoName, withHistory)
			instance = repoData.repo
			go r.registerWebhook(repoName)
		} else {
			instance = repoData.repo
		}
	} else {
		repoData, err = r.clone(repoName, withHistory)
		instance = repoData.repo
		go r.registerWebhook(repoName)
	}
	r.lock.Unlock()

	return instance, err
}

func (r *RepoCache) InstanceForWrite(repoName string) (*git.Repository, string, error) {
	return r.instanceForWrite(repoName, false)
}

func (r *RepoCache) InstanceForWriteWithHistory(repoName string) (*git.Repository, string, error) {
	return r.instanceForWrite(repoName, true)
}

func (r *RepoCache) instanceForWrite(repoName string, withHistory bool) (*git.Repository, string, error) {
	tmpPath, err := ioutil.TempDir("", "gitops-")
	if err != nil {
		errors.WithMessage(err, "couldn't get temporary directory")
	}

	r.lock.Lock()
	if repoData, ok := r.repos[repoName]; ok {
		if withHistory && !repoData.withHistory {
			_, err = r.clone(repoName, withHistory)
			go r.registerWebhook(repoName)
		}
	} else {
		_, err = r.clone(repoName, withHistory)
		go r.registerWebhook(repoName)
	}
	defer r.lock.Unlock()
	if err != nil {
		return nil, "", err
	}

	repoPath := filepath.Join(r.cachePath, strings.ReplaceAll(repoName, "/", "%"))
	err = copy.Copy(repoPath, tmpPath)
	if err != nil {
		errors.WithMessage(err, "could not make copy of repo")
	}

	copiedRepo, err := git.PlainOpen(tmpPath)
	if err != nil {
		return nil, "", fmt.Errorf("cannot open git repository at %s: %s", tmpPath, err)
	}

	return copiedRepo, tmpPath, nil
}

func (r *RepoCache) CleanupWrittenRepo(path string) error {
	return os.RemoveAll(path)
}

func (r *RepoCache) Invalidate(repoName string) {
	r.invalidateCh <- repoName
}

func (r *RepoCache) InvalidateNow(repoName string) {
	logrus.Infof("invalidating repocace now for %s", repoName)
	r.syncGitRepo(repoName)
}

func (r *RepoCache) clone(repoName string, withHistory bool) (repoData, error) {
	if repoName == "" {
		return repoData{}, fmt.Errorf("repo name is mandatory")
	}

	repoPath := filepath.Join(r.cachePath, strings.ReplaceAll(repoName, "/", "%"))

	os.RemoveAll(repoPath)
	err := os.MkdirAll(repoPath, Dir_RWX_RX_R)
	if err != nil {
		return repoData{}, errors.WithMessage(err, "couldn't create folder")
	}

	tokenManager := *r.tokenManager
	token, user, err := tokenManager.Token()
	if err != nil {
		return repoData{}, errors.WithMessage(err, "couldn't get scm token")
	}

	opts := &git.CloneOptions{
		URL: fmt.Sprintf("%s/%s", r.config.ScmURL(), repoName),
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 100,
		Tags:  git.NoTags,
	}
	if withHistory {
		opts.Depth = 0
	}

	repo, err := git.PlainClone(repoPath, false, opts)
	if err != nil {
		return repoData{}, errors.WithMessage(err, "couldn't clone")
	}

	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: FetchRefSpec,
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 100,
		Tags:  git.NoTags,
	})
	if withHistory {
		opts.Depth = 0
	}
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return repoData{}, errors.WithMessage(err, "couldn't fetch")
	}

	r.repos[repoName] = repoData{repo, withHistory}
	return r.repos[repoName], nil
}

func (r *RepoCache) registerWebhook(repoName string) {
	owner, repo := scm.Split(repoName)

	tokenManager := *r.tokenManager
	token, _, err := tokenManager.Token()
	if err != nil {
		logrus.Errorf("couldn't get scm token: %s", err)
	}

	if r.goScmHelper == nil {
		logrus.Warnf("not registering webhook for %s", repoName)
		return
	}
	err = r.goScmHelper.RegisterWebhook(
		r.config.Get(store.Host),
		token,
		r.config.Get(store.WebhookSecret),
		owner,
		repo,
	)
	if err != nil {
		logrus.Warnf("could not register webhook for %s: %s", repoName, err)
	}
}
