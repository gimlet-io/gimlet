package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker/events"
	commonGit "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

const Dir_RWX_RX_R = 0754

var fetchRefSpec = []config.RefSpec{
	"refs/heads/*:refs/heads/*",
}

type BranchDeleteEventWorker struct {
	tokenManager customScm.NonImpersonatedTokenManager
	cachePath    string
	dao          *store.Store
}

func NewBranchDeleteEventWorker(
	tokenManager customScm.NonImpersonatedTokenManager,
	cachePath string,
	dao *store.Store,
) *BranchDeleteEventWorker {
	branchDeleteEventWorker := &BranchDeleteEventWorker{
		tokenManager: tokenManager,
		cachePath:    cachePath,
		dao:          dao,
	}

	return branchDeleteEventWorker
}

func (r *BranchDeleteEventWorker) Run() {
	for {
		reposWithCleanupPolicy, err := r.dao.ReposWithCleanupPolicy()
		if err != nil && err != sql.ErrNoRows {
			logrus.Warnf("could not load repos with cleanup policy: %s", err)
		}
		for _, repoName := range reposWithCleanupPolicy {
			repoPath := filepath.Join(r.cachePath, strings.ReplaceAll(repoName, "/", "%"))
			if _, err := os.Stat(repoPath); err == nil { // repo exist
				repo, err := git.PlainOpen(repoPath)
				if err != nil {
					logrus.Warnf("could not open %s: %s", repoPath, err)
					os.RemoveAll(repoPath)
					continue
				}

				deletedBranches, err := r.detectDeletedBranches(repo)
				if err != nil {
					logrus.Warnf("could not detect deleted branches in %s: %s", repoPath, err)
					os.RemoveAll(repoPath)
					continue
				}
				for _, deletedBranch := range deletedBranches {
					manifests, err := r.extractManifestsFromBranch(repo, deletedBranch)
					if err != nil {
						logrus.Warnf("could not extract manifests: %s", err)
						continue
					}

					branchDeletedEventStr, err := json.Marshal(events.BranchDeletedEvent{
						Repo:      repoName,
						Branch:    deletedBranch,
						Manifests: manifests,
					})
					if err != nil {
						logrus.Warnf("could not serialize branch deleted event: %s", err)
						continue
					}

					// store branch deleted event
					_, err = r.dao.CreateEvent(&model.Event{
						Type:         model.BranchDeletedEvent,
						Blob:         string(branchDeletedEventStr),
						Repository:   repoName,
						GitopsHashes: []string{},
					})
					if err != nil {
						logrus.Warnf("could not store branch deleted event: %s", err)
						continue
					}
				}
			} else if os.IsNotExist(err) {
				err := r.clone(repoName)
				if err != nil {
					logrus.Warnf("could not clone: %s", err)
				}
			} else {
				logrus.Warn(err)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func (r *BranchDeleteEventWorker) detectDeletedBranches(repo *git.Repository) ([]string, error) {
	staleBranches := commonGit.BranchList(repo)

	token, user, err := r.tokenManager.Token()
	if err != nil {
		return []string{}, fmt.Errorf("couldn't get scm token: %s", err)
	}

	err = repo.Fetch(&git.FetchOptions{
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 100,
		Tags:  git.NoTags,
		Prune: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return []string{}, fmt.Errorf("could not fetch: %s", err)
	}

	prunedBranches := commonGit.BranchList(repo)

	return difference(staleBranches, prunedBranches), nil
}

func (r *BranchDeleteEventWorker) extractManifestsFromBranch(repo *git.Repository, branch string) ([]*dx.Manifest, error) {
	var manifests []*dx.Manifest

	files, err := commonGit.RemoteFolderOnBranchWithoutCheckout(repo, branch, ".gimlet")
	if err != nil {
		return manifests, err
	}

	for _, content := range files {
		var mf dx.Manifest
		err = yaml.Unmarshal([]byte(content), &mf)
		if err != nil {
			return manifests, err
		}

		manifests = append(manifests, &mf)
	}

	return manifests, nil
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (r *BranchDeleteEventWorker) clone(repoName string) error {
	repoPath := filepath.Join(r.cachePath, strings.ReplaceAll(repoName, "/", "%"))

	err := os.MkdirAll(repoPath, Dir_RWX_RX_R)
	if err != nil {
		return errors.WithMessage(err, "couldn't create folder")
	}

	token, user, err := r.tokenManager.Token()
	if err != nil {
		os.RemoveAll(repoPath)
		return errors.WithMessage(err, "couldn't get scm token")
	}

	opts := &git.CloneOptions{
		URL: fmt.Sprintf("%s/%s", "https://github.com", repoName),
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 100,
		Tags:  git.NoTags,
	}

	repo, err := git.PlainClone(repoPath, false, opts)
	if err != nil {
		os.RemoveAll(repoPath)
		return errors.WithMessage(err, "couldn't clone")
	}

	err = repo.Fetch(&git.FetchOptions{
		Auth: &http.BasicAuth{
			Username: user,
			Password: token,
		},
		Depth: 1,
		Tags:  git.NoTags,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		os.RemoveAll(repoPath)
		return errors.WithMessage(err, "couldn't fetch")
	}

	return nil
}

func copyRepo(repoPath string) (*git.Repository, error, string) {
	tmpPath, err := ioutil.TempDir("", "gitops-")
	if err != nil {
		return nil, err, ""
	}

	err = copy.Copy(repoPath, tmpPath)
	if err != nil {
		return nil, errors.WithMessage(err, "could not make copy of repo"), tmpPath
	}

	copiedRepo, err := git.PlainOpen(tmpPath)
	return copiedRepo, err, tmpPath
}
