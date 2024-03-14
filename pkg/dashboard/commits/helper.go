package commits

import (
	"fmt"
	"strings"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
)

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
