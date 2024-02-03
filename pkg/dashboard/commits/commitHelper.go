package commits

import (
	"fmt"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
		err = generateFakeArtifactsForCommits(hashes, branch, dao, repo, repoName)
	})
	if err != nil {
		return err
	}

	return nil
}

func AssureSCMData(
	repo string,
	hashes []string,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
) error {
	return assureSCMDataWithRetry(
		repo,
		hashes,
		dao,
		gitServiceImpl,
		token,
		false,
	)
}

func assureSCMDataWithRetry(
	repo string,
	hashes []string,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
	isRetry bool,
) error {
	dbCommits, err := dao.CommitsByRepoAndSHA(repo, hashes)
	if err != nil {
		return fmt.Errorf("cannot get commits from db %s", err)
	}

	dbCommitsByHash := map[string]*model.Commit{}
	for _, dbCommit := range dbCommits {
		dbCommitsByHash[dbCommit.SHA] = dbCommit
	}

	var hashesToFetch []string
	for _, hash := range hashes {
		if _, ok := dbCommitsByHash[hash]; !ok {
			hashesToFetch = append(hashesToFetch, hash)
		}
	}

	// if not all commit was fetched already fetch them
	if len(hashesToFetch) > 0 && !isRetry {
		logrus.Infof("Fetching scm data for %s", strings.Join(hashesToFetch, ","))
		// fetch remote commit info, then try to assure again
		owner, name := scm.Split(repo)
		fetchCommits(owner, name, gitServiceImpl, token, dao, hashesToFetch)
		return assureSCMDataWithRetry(
			repo,
			hashes,
			dao,
			gitServiceImpl,
			token,
			true,
		)
	}

	return nil
}

// commits come from go-scm based live git traversal
// we decorate that commit data with data from SCMs with the
// following fields: url, author, author_pic, message, created, tags, status
// URL, Author, AuthorPic, Tags, Status
func fetchCommits(
	owner string,
	repo string,
	gitService customScm.CustomGitService,
	token string,
	store *store.Store,
	hashesToFetch []string,
) {
	if len(hashesToFetch) == 0 {
		return
	}

	commits, err := gitService.FetchCommits(owner, repo, token, hashesToFetch)
	if err != nil {
		logrus.Errorf("Could not fetch commits for %v, %v", repo, err)
		return
	}

	err = store.SaveCommits(scm.Join(owner, repo), commits)
	if err != nil {
		logrus.Errorf("Could not store commits for %v, %v", repo, err)
		return
	}
	statusOnCommits := map[string]*model.CombinedStatus{}
	for _, c := range commits {
		statusOnCommits[c.SHA] = &c.Status
	}

	if len(statusOnCommits) != 0 {
		err = store.SaveStatusesOnCommits(scm.Join(owner, repo), statusOnCommits)
		if err != nil {
			logrus.Errorf("Could not store status for %v, %v", repo, err)
			return
		}
	}
}

func generateFakeArtifactsForCommits(hashes []string, branch string, store *store.Store, repo *git.Repository, repoName string) error {
	for _, hash := range hashes {
		key := fmt.Sprintf("%s-%s", model.CommitArtifactsGenerated, hash)
		_, err := store.KeyValue(key)
		if err == nil {
			continue
		}

		err = generateFakeArtifact(hash, branch, store, repo, repoName)
		if err != nil {
			return err
		}

		store.SaveKeyValue(&model.KeyValue{
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
			"SHA": hash,
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
