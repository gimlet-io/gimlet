package notifications

import (
	"fmt"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/xanzy/go-gitlab"
)

const gitlabCommitLink = "%s/%s/-/commit/%s"

type gitlabProvider struct {
	tokenManager customScm.NonImpersonatedTokenManager
	baseUrl      string
}

func NewGitlabProvider(
	tokenManager customScm.NonImpersonatedTokenManager,
	baseUrl string,
) *gitlabProvider {
	return &gitlabProvider{
		tokenManager: tokenManager,
		baseUrl:      baseUrl,
	}
}

func (g *gitlabProvider) send(msg Message) error {
	status, err := msg.AsStatus()
	if err != nil {
		return fmt.Errorf("cannot create github status message: %s", err)
	}

	if status == nil {
		return nil
	}

	repositoryName := msg.RepositoryName()
	parts := strings.Split(repositoryName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("cannot determine repo owner and name")
	}
	owner := parts[0]
	repo := parts[1]

	sha := msg.SHA()

	return g.post(owner, repo, sha, status)
}

func (g *gitlabProvider) post(owner string, repo string, sha string, status *status) error {
	token, _, _ := g.tokenManager.Token()
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(g.baseUrl))
	if err != nil {
		return fmt.Errorf("couldn't create gitlab client: %s", err)
	}

	targetURL := fmt.Sprintf(gitlabCommitLink, g.baseUrl, status.repo, status.sha)
	if status.state == "failure" {
		targetURL = ""
	}

	_, _, err = git.Commits.SetCommitStatus(
		scm.Join(owner, repo),
		sha,
		&gitlab.SetCommitStatusOptions{
			State:       gitlab.BuildStateValue(status.state),
			Name:        &status.context,
			Context:     &status.context,
			TargetURL:   &targetURL,
			Description: &status.description,
		},
	)

	return err
}
