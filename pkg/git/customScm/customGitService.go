package customScm

import (
	"context"

	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet/pkg/git/customScm/customGitlab"
	"github.com/google/go-github/v37/github"
)

const (
	BodyProcessing = `### <span aria-hidden="true">👷</span> Deploy Preview for *%s* processing.

| Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
`

	BodyReady = `### <span aria-hidden="true">✅</span> Deploy Preview for *%s* ready!

| Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
|<span aria-hidden="true">😎</span> Deploy Preview | [https://%s](https://%s) |
`

	BodyFailed = `### <span aria-hidden="true">❌</span> Deploy Preview for *%s* failed.

|  Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
`
)

type CustomGitService interface {
	FetchCommits(owner string, repo string, token string, hashesToFetch []string) ([]*model.Commit, error)
	InstallationRepos(installationToken string) ([]string, error)
	GetAppNameAndAppSettingsURLs(installationToken string, ctx context.Context) (string, string, string, error)
	CreateRepository(owner string, name string, loggedInUser string, orgToken string, token string) error
	AddDeployKeyToRepo(owner, repo, token, keyTitle, keyValue string, canWrite bool) error
	Comments(token, repoName string, pullNumber int) ([]*github.IssueComment, error)
	CreateComment(token, repoName string, pullNumber int, body string) error
	UpdateComment(token, repoName string, commentId int64, body string) error
}

func NewGitService(dynamicConfig *dynamicconfig.DynamicConfig) CustomGitService {
	var gitSvc CustomGitService

	if dynamicConfig.IsGithub() {
		gitSvc = &customGithub.GithubClient{}
	} else if dynamicConfig.IsGitlab() {
		gitSvc = &customGitlab.GitlabClient{
			BaseURL: dynamicConfig.ScmURL(),
		}
	} else {
		gitSvc = &DummyGitService{}
	}
	return gitSvc
}

type DummyGitService struct {
}

func (d *DummyGitService) FetchCommits(owner string, repo string, token string, hashesToFetch []string) ([]*model.Commit, error) {
	return nil, nil
}

func (d *DummyGitService) InstallationRepos(installationToken string) ([]string, error) {
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

func (d *DummyGitService) Comments(token, repoName string, pullNumber int) ([]*github.IssueComment, error) {
	return nil, nil
}

func (d *DummyGitService) CreateComment(token, repoName string, pullNumber int, body string) error {
	return nil
}

func (d *DummyGitService) UpdateComment(token, repoName string, commentId int64, body string) error {
	return nil
}
