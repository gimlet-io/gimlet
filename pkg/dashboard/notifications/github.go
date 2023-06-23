package notifications

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/google/go-github/v37/github"
	githubLib "github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

const githubCommitLink = "https://github.com/%s/commit/%s"

type githubProvider struct {
	tokenManager *customScm.NonImpersonatedTokenManager
}

func NewGithubProvider(tokenManager *customScm.NonImpersonatedTokenManager) *githubProvider {
	return &githubProvider{
		tokenManager: tokenManager,
	}
}

func (g *githubProvider) send(msg Message) error {
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

	targetURL := fmt.Sprintf(githubCommitLink, status.repo, status.sha)
	if status.state == "failure" {
		targetURL = ""
	}

	urlPtr := &targetURL
	if targetURL == "" {
		urlPtr = nil
	}

	return g.post(owner, repo, sha, &githubLib.RepoStatus{
		State:       &status.state,
		Context:     &status.context,
		Description: &status.description,
		TargetURL:   urlPtr,
	})
}

func (g *githubProvider) post(owner string, repo string, sha string, status *githubLib.RepoStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tokenManager := *g.tokenManager
	token, _, err := tokenManager.Token()
	if err != nil {
		return fmt.Errorf("couldn't get scm token: %s", err)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opts := &github.ListOptions{PerPage: 50}
	statuses, _, err := client.Repositories.ListStatuses(ctx, owner, repo, sha, opts)
	if err != nil {
		return fmt.Errorf("could not list commit statuses: %v", err)
	}
	if statusExists(statuses, status) {
		return nil
	}

	_, _, err = client.Repositories.CreateStatus(ctx, owner, repo, sha, status)
	if err != nil {
		return fmt.Errorf("could not create commit status: %v", err)
	}

	return nil
}

func statusExists(statuses []*github.RepoStatus, status *github.RepoStatus) bool {
	for _, s := range statuses {
		if *s.Context == *status.Context {
			if *s.State == *status.State && *s.Description == *status.Description {
				return true
			}

			return false
		}
	}

	return false
}
