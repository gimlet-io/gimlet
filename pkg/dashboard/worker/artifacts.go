package worker

import (
	"fmt"
	"slices"
	"time"

	"github.com/gimlet-io/gimlet/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/uuid"
)

type ArtifactsWorker struct {
	gitRepoCache *nativeGit.RepoCache
	dao          *store.Store
	trigger      chan string
	gitopsQueue  chan int
}

func NewArtifactsWorker(
	gitRepoCache *nativeGit.RepoCache,
	dao *store.Store,
	trigger chan string,
	gitopsQueue chan int,
) *ArtifactsWorker {
	return &ArtifactsWorker{gitRepoCache: gitRepoCache, dao: dao, trigger: trigger, gitopsQueue: gitopsQueue}
}

func (a *ArtifactsWorker) Run() {
	for {
		repoName := <-a.trigger
		go a.assureGimletArtifacts(repoName)
	}
}

func (a *ArtifactsWorker) assureGimletArtifacts(repoName string) error {
	err := a.gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
		var innerErr error
		headBranch, innerErr := nativeGit.HeadBranch(repo) //TODO we need to generate this for all branches wehn it comes to preview apps
		if innerErr != nil {
			return innerErr
		}

		hashes, innerErr := lastTenCommits(repo, headBranch)
		if innerErr != nil {
			return innerErr
		}

		slices.Reverse(hashes) //artifacts should be generated in commit creation order

		err := generateFakeArtifactsForCommits(repoName, headBranch, hashes, a.dao, repo)
		a.gitopsQueue <- 1

		return err
	})

	return err
}

func lastTenCommits(repo *git.Repository, branch string) ([]string, error) {
	branchHeadHash := nativeGit.BranchHeadHash(repo, branch)

	commitWalker, err := repo.Log(&git.LogOptions{
		From: branchHeadHash,
	})
	if err != nil {
		return nil, err
	}

	limit := 10
	hashes := []string{}
	err = commitWalker.ForEach(func(c *object.Commit) error {
		if limit != 0 && len(hashes) >= limit {
			return fmt.Errorf("%s", "LIMIT")
		}

		hashes = append(hashes, c.Hash.String())
		return nil
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "LIMIT" {
		return nil, err
	}

	return hashes, nil
}

func generateFakeArtifactsForCommits(
	repoName string,
	branch string,
	hashes []string,
	dao *store.Store,
	repo *git.Repository,
) error {
	for _, hash := range hashes {
		key := fmt.Sprintf("%s-%s", model.CommitArtifactsGenerated, hash)
		_, err := dao.KeyValue(key)
		if err == nil {
			continue
		}

		err = generateFakeArtifact(hash, branch, dao, repo, repoName)
		if err != nil {
			return err
		}

		dao.SaveKeyValue(&model.KeyValue{
			Key: key,
		})
	}

	return nil
}

func generateFakeArtifact(hash string, branch string, store *store.Store, repo *git.Repository, repoName string) error {
	manifests, err := gitops.Manifests(repo, hash)
	if err != nil {
		return err
	}

	manifestsThatNeedFakeArtifact := []*dx.Manifest{}
	for _, m := range manifests {
		strategy := gitops.ExtractImageStrategy(m)

		if strategy == "static" ||
			strategy == "static-site" ||
			strategy == "buildpacks" ||
			strategy == "dockerfile" {
			manifestsThatNeedFakeArtifact = append(manifestsThatNeedFakeArtifact, m)
		}
	}

	if len(manifestsThatNeedFakeArtifact) == 0 {
		return nil
	}

	err = doGenerateFakeArtifact(hash, branch, manifestsThatNeedFakeArtifact, store, repoName, repo)
	if err != nil {
		return err
	}

	return nil
}

func doGenerateFakeArtifact(
	hash string,
	branch string,
	manifests []*dx.Manifest,
	store *store.Store,
	repoName string,
	repo *git.Repository,
) error {
	owner, name := scm.Split(repoName)
	version, err := gitops.Version(owner, name, repo, hash, branch)
	if err != nil {
		return err
	}

	artifact := &dx.Artifact{
		ID:           fmt.Sprintf("%s-%s", repoName, uuid.New().String()),
		Created:      time.Now().Unix(),
		Fake:         true,
		Environments: manifests,
		Version:      *version,
		Vars: map[string]string{
			"SHA":    hash,
			"REPO":   repoName,
			"OWNER":  owner,
			"BRANCH": branch,
		},
	}

	event, err := model.ToEvent(*artifact)
	if err != nil {
		return err
	}

	_, err = store.CreateEvent(event)
	if err != nil {
		return err
	}

	return nil
}
