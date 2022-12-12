package customGitlab

import (
	"context"
	"fmt"
	"sync"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
}

// FetchCommits gets commit data that we use to decorate
// the go-scm based commit info we have. Fields we fetch:
// URL, Author, AuthorPic, Tags, Status
func (c *GitlabClient) FetchCommits(
	owner string,
	repo string,
	token string,
	hashesToFetch []string,
) ([]*model.Commit, error) {
	git, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}

	commits := make(chan *model.Commit, 5)
	waitCh := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for _, hash := range hashesToFetch {
			wg.Add(1)

			go func(sha string, commits chan *model.Commit) {
				defer wg.Done()
				statuses, _, err := git.Commits.GetCommitStatuses(
					scm.Join(owner, repo),
					sha,
					&gitlab.GetCommitStatusesOptions{},
				)
				if err != nil {
					logrus.Warnf("couldn't fetch commit info (%s): %s", sha, err)
				}
				fmt.Println(statuses)
				commits <- &model.Commit{}
			}(hash, commits)
		}

		wg.Wait()
		close(waitCh)
	}()

	fetched := []*model.Commit{}
	select {
	case <-waitCh:
	case c := <-commits:
		fetched = append(fetched, c)
	}

	return fetched, nil
}

func (c *GitlabClient) OrgRepos(installationToken string) ([]string, error) {
	return nil, nil
}

func (c *GitlabClient) GetAppNameAndAppSettingsURLs(appToken string, ctx context.Context) (string, string, string, error) {
	return "", "", "", nil
}

func (c *GitlabClient) CreateRepository(owner string, repo string, loggedInUser string, orgToken string, userToken string) error {
	git, err := gitlab.NewClient(orgToken)
	if err != nil {
		return err
	}

	_, _, err = git.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name: &repo,
	})
	return err
}
