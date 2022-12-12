package customGitlab

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
}

func (c *GitlabClient) FetchCommits(
	owner string,
	repo string,
	token string,
	hashesToFetch []string,
) ([]*model.Commit, error) {
	return nil, nil
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
