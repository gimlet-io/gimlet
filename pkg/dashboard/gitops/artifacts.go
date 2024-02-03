package gitops

import (
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
)

// factor new commit routine and decoration out of a client facing api
// needs to be called from hooks as well
// and every place we get to know about new commits
// - should not trigger policy, if there is a newer artifact - geezus
func AssureGimletArtifacts(
	repoName string,
	branch string,
	hashes []string,
	gitRepoCache *nativeGit.RepoCache,
	dao *store.Store,
) error {
	var err error
	gitRepoCache.PerformAction(repoName, func(repo *git.Repository) {
		err = generateFakeArtifactsForCommits(repoName, branch, hashes, dao, repo)
	})
	if err != nil {
		return err
	}

	return nil
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
	manifests, err := Manifests(repo, hash)
	if err != nil {
		return err
	}

	manifestsThatNeedFakeArtifact := []*dx.Manifest{}
	for _, m := range manifests {
		strategy := ExtractImageStrategy(m)

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
	version, err := Version(owner, name, repo, hash, branch)
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
