package customGitlab

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
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
