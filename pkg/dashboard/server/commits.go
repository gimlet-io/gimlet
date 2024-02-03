package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-chi/chi"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

func commits(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	repoName := fmt.Sprintf("%s/%s", owner, name)
	branch := r.URL.Query().Get("branch")
	hashString := r.URL.Query().Get("fromHash")

	ctx := r.Context()
	dao := ctx.Value("store").(*store.Store)
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	var err error
	if branch == "" {
		gitRepoCache.PerformAction(repoName, func(repo *git.Repository) {
			branch, err = helper.HeadBranch(repo)
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logrus.Errorf("cannot get head branch: %s", err)
			return
		}
	}

	var hash plumbing.Hash
	gitRepoCache.PerformAction(repoName, func(repo *git.Repository) {
		hash = helper.BranchHeadHash(repo, branch)
	})

	if hashString != "head" {
		hash = plumbing.NewHash(hashString)
	}

	var commitWalker object.CommitIter
	gitRepoCache.PerformAction(repoName, func(repo *git.Repository) {
		commitWalker, err = repo.Log(&git.LogOptions{
			From: hash,
		})
	})
	if err != nil {
		logrus.Errorf("cannot walk commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	limit := 10
	commits := []*Commit{}
	hashes := []string{}
	err = commitWalker.ForEach(func(c *object.Commit) error {
		if limit != 0 && len(commits) >= limit {
			return fmt.Errorf("%s", "LIMIT")
		}

		commits = append(commits, &Commit{
			SHA:        c.Hash.String(),
			AuthorName: c.Author.Name,
			Message:    c.Message,
			CreatedAt:  c.Author.When.Unix(),
		})
		hashes = append(hashes, c.Hash.String())

		return nil
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "LIMIT" {
		logrus.Errorf("cannot walk commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	commits, err = decorateWithSCMData(repoName, commits, dao)
	if err != nil {
		logrus.Errorf("cannot decorate commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	commits, err = decorateWithDeployTargets(commits, dao)
	if err != nil {
		logrus.Errorf("cannot decorate commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	commits = squashCommitStatuses(commits)

	commitsString, err := json.Marshal(commits)
	if err != nil {
		logrus.Errorf("cannot serialize commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(commitsString)
}

func decorateWithSCMData(repoName string, commits []*Commit, dao *store.Store) ([]*Commit, error) {
	hashes := []string{}
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	dbCommits, err := dao.CommitsByRepoAndSHA(repoName, hashes)
	if err != nil {
		return nil, err
	}

	dbCommitsByHash := map[string]*model.Commit{}
	for _, dbCommit := range dbCommits {
		dbCommitsByHash[dbCommit.SHA] = dbCommit
	}

	for _, commit := range commits {
		dbCommit := dbCommitsByHash[commit.SHA]
		commit.URL = dbCommit.URL
		commit.Author = dbCommit.Author
		commit.AuthorPic = dbCommit.AuthorPic
		commit.Tags = dbCommit.Tags
		commit.Status = dbCommit.Status
	}

	return commits, nil
}

func decorateWithDeployTargets(commits []*Commit, store *store.Store) ([]*Commit, error) {
	hashes := []string{}
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	events, err := store.Artifacts(
		"", "",
		nil,
		"",
		hashes,
		500, 0, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get artifacts: %s", err)
	}

	artifacts := []*dx.Artifact{}
	for _, a := range events {
		artifact, err := model.ToArtifact(a)
		if err != nil {
			return nil, fmt.Errorf("cannot deserialize artifact: %s", err)
		}
		artifacts = append(artifacts, artifact)
	}

	artifactsBySha := map[string][]*dx.Artifact{}
	for _, a := range artifacts {
		if artifactsBySha[a.Version.SHA] == nil {
			artifactsBySha[a.Version.SHA] = []*dx.Artifact{}
		}
		artifactsBySha[a.Version.SHA] = append(artifactsBySha[a.Version.SHA], a)
	}

	var decoratedCommits []*Commit
	for _, c := range commits {
		if as, ok := artifactsBySha[c.SHA]; ok {
			for _, artifact := range as {
				for _, targetEnv := range artifact.Environments {
					targetEnv.ResolveVars(artifact.CollectVariables())
					if c.DeployTargets == nil {
						c.DeployTargets = []*api.DeployTarget{}
					}
					if deployTargetExists(c.DeployTargets, targetEnv.App, targetEnv.Env) {
						continue
					}
					c.DeployTargets = append(c.DeployTargets, &api.DeployTarget{
						App:        targetEnv.App,
						Env:        targetEnv.Env,
						Tenant:     targetEnv.Tenant.Name,
						ArtifactId: artifact.ID,
					})
				}
			}
		}
		decoratedCommits = append(decoratedCommits, c)
	}

	return decoratedCommits, nil
}

func deployTargetExists(targets []*api.DeployTarget, app string, env string) bool {
	for _, t := range targets {
		if t.App == app && t.Env == env {
			return true
		}
	}

	return false
}

// Commit represents a Github commit
type Commit struct {
	SHA           string               `json:"sha"`
	URL           string               `json:"url"`
	Author        string               `json:"author"`
	AuthorName    string               `json:"authorName"`
	AuthorPic     string               `json:"author_pic"`
	Message       string               `json:"message"`
	CreatedAt     int64                `json:"created_at"`
	Tags          []string             `json:"tags,omitempty"`
	Status        model.CombinedStatus `json:"status,omitempty"`
	DeployTargets []*api.DeployTarget  `json:"deployTargets,omitempty"`
}

func squashCommitStatuses(commits []*Commit) []*Commit {
	var commitsWithSquashedStatuses []*Commit

	for _, commit := range commits {
		statusMap := map[string]model.CommitStatus{}
		for _, s := range commit.Status.Contexts {
			// Statuses are returned in reverse chronological order
			// we only keep the latest
			if _, ok := statusMap[s.Context]; ok {
				continue
			}

			statusMap[s.Context] = s
		}

		commit.Status.Contexts = []model.CommitStatus{}
		for _, status := range statusMap {
			commit.Status.Contexts = append(commit.Status.Contexts, status)
		}

		commitsWithSquashedStatuses = append(commitsWithSquashedStatuses, commit)
	}

	return commitsWithSquashedStatuses
}

func triggerCommitSync(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	repoName := fmt.Sprintf("%s/%s", owner, name)

	ctx := r.Context()
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	gitRepoCache.Invalidate(repoName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
