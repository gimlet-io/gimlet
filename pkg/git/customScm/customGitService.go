package customScm

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type CustomGitService interface {
	FetchCommits(owner string, repo string, token string, hashesToFetch []string) ([]*model.Commit, error)
	OrgRepos(installationToken string) ([]string, error)
	GetAppNameAndAppSettingsURLs(installationToken string, ctx context.Context) (string, string, string, error)
	CreateRepository(owner string, name string, loggedInUser string, orgToken string, token string) error
	AddDeployKeyToRepo(orgToken, userToken, repo, loggedInUser, keyTitle, keyValue string, readOnly bool) error
}
