package customGitlab

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
	BaseURL string
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
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://"+c.BaseURL))
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

				commit, _, err := git.Commits.GetCommit(scm.Join(owner, repo), sha)
				if err != nil {
					logrus.Warnf("couldn't fetch commit info (%s): %s", sha, err)
					return
				}

				statuses, _, err := git.Commits.GetCommitStatuses(
					scm.Join(owner, repo),
					sha,
					&gitlab.GetCommitStatusesOptions{},
				)
				if err != nil {
					logrus.Warnf("couldn't fetch commit status (%s): %s", sha, err)
				}

				contexts := []model.Status{}
				for _, s := range statuses {
					state := fromGitlabStatus(s.Status)
					contexts = append(contexts, model.Status{
						State:       state,
						Context:     s.Name,
						CreatedAt:   s.CreatedAt.Format(time.RFC3339),
						TargetUrl:   s.TargetURL,
						Description: s.Description,
					})
				}

				var commitStatus string
				if commit.Status != nil {
					commitStatus = string(*commit.Status)
				}

				commits <- &model.Commit{
					SHA:     sha,
					Message: commit.Message,
					URL:     commit.WebURL,
					Status: model.CombinedStatus{
						State:    commitStatus,
						Contexts: contexts,
					},
				}
			}(hash, commits)
		}

		wg.Wait()
		close(waitCh)
	}()

	fetched := []*model.Commit{}
	for {
		select {
		case <-waitCh:
			return fetched, nil
		case c := <-commits:
			fetched = append(fetched, c)
		}
	}
}

func (c *GitlabClient) OrgRepos(token string) ([]string, error) {
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://"+c.BaseURL))
	if err != nil {
		return nil, err
	}

	projects, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{})
	if err != nil {
		return nil, err
	}
	repos := []string{}
	for _, p := range projects {
		repos = append(repos, p.Name)
	}

	return repos, nil
}

func (c *GitlabClient) GetAppNameAndAppSettingsURLs(appToken string, ctx context.Context) (string, string, string, error) {
	return "", "", "", nil
}

func (c *GitlabClient) CreateRepository(owner string, repo string, loggedInUser string, orgToken string, userToken string) error {
	git, err := gitlab.NewClient(orgToken, gitlab.WithBaseURL("https://"+c.BaseURL))
	if err != nil {
		return err
	}

	var namespaceId int
	groups, _, err := git.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return err
	}
	for _, g := range groups {
		if g.Name == owner {
			namespaceId = g.ID
		}
	}

	if namespaceId != 0 {
		_, _, err = git.Projects.CreateProject(&gitlab.CreateProjectOptions{
			Name:        &repo,
			NamespaceID: &namespaceId,
		})
	} else {
		_, _, err = git.Projects.CreateProject(&gitlab.CreateProjectOptions{
			Name: &repo,
		})
	}

	return err
}

func fromGitlabStatus(gitlabStatus string) string {
	// https://docs.gitlab.com/ee/api/commits.html#set-the-pipeline-status-of-a-commit
	// https://github.com/gimlet-io/gimlet/blob/1997f9f8f08ccff96828b239b5126632b47dee77/web/dashboard/src/components/commits/commits.js#L183
	switch gitlabStatus {
	case "running":
		return "IN_PROGRESS"
	default:
		return strings.ToUpper(gitlabStatus)
	}
}
