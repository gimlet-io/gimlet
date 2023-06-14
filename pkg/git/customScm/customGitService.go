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
	AddDeployKeyToRepo(owner, repo, token, keyTitle, keyValue string, canWrite bool) error
}

type DummyGitService struct {
}

func NewDummyGitService() *DummyGitService {
	return &DummyGitService{}
}

func (d *DummyGitService) FetchCommits(owner string, repo string, token string, hashesToFetch []string) ([]*model.Commit, error) {
	return nil, nil
}

func (d *DummyGitService) OrgRepos(installationToken string) ([]string, error) {
	return nil, nil
}

func (d *DummyGitService) GetAppNameAndAppSettingsURLs(installationToken string, ctx context.Context) (string, string, string, error) {
	return "", "", "", nil
}

func (d *DummyGitService) CreateRepository(owner string, name string, loggedInUser string, orgToken string, token string) error {
	return nil
}

func (d *DummyGitService) AddDeployKeyToRepo(owner, repo, token, keyTitle, keyValue string, canWrite bool) error {
	return nil
}
