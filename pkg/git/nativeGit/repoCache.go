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
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
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
	tokenManager customScm.NonImpersonatedTokenManager
	repos        map[string]*repoData
	reposMapLock sync.Mutex // lock this if you add or remove items from the repos map
	stopCh       chan struct{}

	// For webhook registration
	config        *dashboardConfig.Config
	dynamicConfig *dynamicconfig.DynamicConfig
	clientHub     *streaming.ClientHub

	// for builtin env
	gitUser *model.User

	// to hydrate commits
	triggerArtifactGeneration chan string
}

type repoData struct {
	repo        *git.Repository
	withHistory bool
	lock        sync.Mutex // lock this if you do git operation on this repo
}

const BRANCH_DELETED_WORKER_SUBPATH = "branch-deleted-worker"

func NewRepoCache(
	tokenManager customScm.NonImpersonatedTokenManager,
	stopCh chan struct{},
	config *dashboardConfig.Config,
	dynamicConfig *dynamicconfig.DynamicConfig,
	clientHub *streaming.ClientHub,
	gitUser *model.User,
	triggerArtifactGeneration chan string,
) (*RepoCache, error) {
	repoCache := &RepoCache{
		tokenManager:              tokenManager,
		repos:                     map[string]*repoData{},
		stopCh:                    stopCh,
		config:                    config,
		dynamicConfig:             dynamicConfig,
		clientHub:                 clientHub,
		gitUser:                   gitUser,
		triggerArtifactGeneration: triggerArtifactGeneration,
	}

	const DirRwxRxR = 0754
	cachePath := config.RepoCachePath
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
		if fileInfo.Name() == BRANCH_DELETED_WORKER_SUBPATH {
			continue
		}
		if fileInfo.Name() == "lost+found" {
			continue
		}

		path := filepath.Join(cachePath, fileInfo.Name())
		repo, err := git.PlainOpen(path)
		if err != nil {
			logrus.Warnf("cannot open git repository at %s: %s", path, err)
			continue
		}

		repoCache.repos[strings.ReplaceAll(fileInfo.Name(), "%", "/")] = &repoData{repo: repo, withHistory: false}
	}

	return repoCache, nil
}

func (r *RepoCache) Run() {
	for {
		t0 := time.Now()
		for repoName, _ := range r.repos {
			r.syncGitRepo(repoName)
		}
		logrus.Debugf("Synching repos took %f seconds", time.Since(t0).Seconds())

		select {
		case <-r.stopCh:
			logrus.Info("stopping")
			return
		case <-time.After(30 * time.Second):
		}
	}
}

func (r *RepoCache) syncGitRepo(repoName string) {
	var auth *http.BasicAuth
	owner, _ := scm.Split(repoName)
	if owner == "builtin" {
		auth = &http.BasicAuth{
			Username: r.gitUser.Login,
			Password: r.gitUser.Secret,
		}
	} else {
		token, _, err := r.tokenManager.Token()
		if err != nil {
			logrus.Errorf("couldn't get scm token: %s", err)
			return
		}
		auth = &http.BasicAuth{
			Username: "123",
			Password: token,
		}
	}

	if _, ok := r.repos[repoName]; !ok {
		logrus.Warnf("could not get repo by name from cache: %s", repoName)
		return // preventing a race condition in cleanup
	}

	repoData := r.repos[repoName]
	repo := repoData.repo

	opts := &git.FetchOptions{
		RefSpecs: FetchRefSpec,
		Auth:     auth,
		Depth:    100,
		Tags:     git.NoTags,
		Prune:    true,
	}
	if repoData.withHistory {
		opts.Depth = 0
	}

	repoData.lock.Lock()
	defer repoData.lock.Unlock()

	err := repo.Fetch(opts)
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

	r.triggerArtifactGeneration <- repoName

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
	r.reposMapLock.Lock()
	delete(r.repos, repoName)
	r.reposMapLock.Unlock()
}

func (r *RepoCache) PerformAction(repoName string, fn func(repo *git.Repository) error) error {
	repo, err := r.instanceForRead(repoName, false)
	if err != nil {
		return err
	}

	repo.lock.Lock()
	err = fn(repo.repo)
	repo.lock.Unlock()

	return err
}

func (r *RepoCache) PerformActionWithHistory(repoName string, fn func(repo *git.Repository)) error {
	repo, err := r.instanceForRead(repoName, true)
	repo.lock.Lock()
	fn(repo.repo)
	repo.lock.Unlock()

	return err
}

func (r *RepoCache) instanceForRead(repoName string, withHistory bool) (repo *repoData, err error) {
	var repoData repoData
	if existingRepoData, ok := r.repos[repoName]; ok {
		if withHistory && !existingRepoData.withHistory {
			repoData, err = r.clone(repoName, withHistory)
			repo = &repoData
			go r.registerWebhook(repoName)
		} else {
			repo = existingRepoData
		}
	} else {
		repoData, err = r.clone(repoName, withHistory)
		repo = &repoData
		go r.registerWebhook(repoName)
	}

	return repo, err
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

	if repoData, ok := r.repos[repoName]; ok {
		if withHistory && !repoData.withHistory {
			_, err = r.clone(repoName, withHistory)
			go r.registerWebhook(repoName)
		}
	} else {
		_, err = r.clone(repoName, withHistory)
		go r.registerWebhook(repoName)
	}
	if err != nil {
		return nil, "", err
	}

	repoData := r.repos[repoName]
	repoData.lock.Lock()
	defer repoData.lock.Unlock()

	repoPath := filepath.Join(r.config.RepoCachePath, strings.ReplaceAll(repoName, "/", "%"))
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
	logrus.Debugf("invalidating repocache for %s", repoName)
	r.syncGitRepo(repoName)
}

func (r *RepoCache) clone(repoName string, withHistory bool) (repoData, error) {
	if repoName == "" {
		return repoData{}, fmt.Errorf("repo name is mandatory")
	}

	repoPath := filepath.Join(r.config.RepoCachePath, strings.ReplaceAll(repoName, "/", "%"))

	r.reposMapLock.Lock()
	defer r.reposMapLock.Unlock()
	os.RemoveAll(repoPath)
	err := os.MkdirAll(repoPath, Dir_RWX_RX_R)
	if err != nil {
		return repoData{}, errors.WithMessage(err, "couldn't create folder")
	}

	var auth *http.BasicAuth
	var url string
	owner, _ := scm.Split(repoName)
	if owner == "builtin" {
		url = fmt.Sprintf("http://%s/%s", r.config.GitHost, repoName)
		auth = &http.BasicAuth{
			Username: r.gitUser.Login,
			Password: r.gitUser.Secret,
		}
	} else {
		url = fmt.Sprintf("%s/%s", r.dynamicConfig.ScmURL(), repoName)
		token, _, err := r.tokenManager.Token()
		if err != nil {
			return repoData{}, errors.WithMessage(err, "couldn't get scm token")
		}
		auth = &http.BasicAuth{
			Username: "123",
			Password: token,
		}
	}

	opts := &git.CloneOptions{
		URL:   url,
		Auth:  auth,
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
		Auth:     auth,
		Depth:    100,
		Tags:     git.NoTags,
	})
	if withHistory {
		opts.Depth = 0
	}
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return repoData{}, errors.WithMessage(err, "couldn't fetch")
	}

	r.repos[repoName] = &repoData{repo: repo, withHistory: withHistory}
	return repoData{repo: repo, withHistory: withHistory}, nil
}

func (r *RepoCache) registerWebhook(repoName string) {
	owner, repo := scm.Split(repoName)

	if owner == "builtin" {
		return
	}

	token, _, err := r.tokenManager.Token()
	if err != nil {
		logrus.Errorf("couldn't get scm token: %s", err)
	}

	goScmHelper := genericScm.NewGoScmHelper(r.dynamicConfig, nil)
	err = goScmHelper.RegisterWebhook(
		r.config.Host,
		token,
		r.config.WebhookSecret,
		owner,
		repo,
	)
	if err != nil {
		logrus.Warnf("could not register webhook for %s: %s", repoName, err)
	}
}
