package customScm

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm/customGitlab"
)

type CustomGitService interface {
	FetchCommits(owner string, repo string, token string, hashesToFetch []string) ([]*model.Commit, error)
	OrgRepos(installationToken string) ([]string, error)
	GetAppNameAndAppSettingsURLs(installationToken string, ctx context.Context) (string, string, string, error)
	CreateRepository(owner string, name string, loggedInUser string, orgToken string, token string) error
	AddDeployKeyToRepo(owner, repo, token, keyTitle, keyValue string, canWrite bool) error
}

func NewGitService(config *config.Config) CustomGitService {
	var gitSvc CustomGitService

	if config.IsGithub() {
		gitSvc = &customGithub.GithubClient{}
	} else if config.IsGitlab() {
		gitSvc = &customGitlab.GitlabClient{
			BaseURL: config.ScmURL(),
		}
	}
	return gitSvc
}
