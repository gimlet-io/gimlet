package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/api"
	commitsHelper "github.com/gimlet-io/gimlet/pkg/dashboard/commits"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/go-chi/chi/v5"
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

	if branch == "" {
		err := gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
			var err error
			branch, err = nativeGit.HeadBranch(repo)
			return err
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logrus.Errorf("cannot get head branch: %s", err)
			return
		}
	}

	var hash plumbing.Hash
	if hashString != "head" {
		hash = plumbing.NewHash(hashString)
	} else {
		gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
			logrus.Debugf("getting branchheadhash for %s", branch)
			hash = nativeGit.BranchHeadHash(repo, branch)
			return nil
		})
	}

	var commitWalker object.CommitIter
	var err error
	err = gitRepoCache.PerformAction(repoName, func(repo *git.Repository) error {
		var err error
		commitWalker, err = repo.Log(&git.LogOptions{
			From: hash,
		})
		return err
	})
	if err != nil {
		logrus.Errorf("cannot walk commits from %s: %s", hash, err)
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

	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	err = commitsHelper.AssureSCMData(repoName, hashes, dao, gitServiceImpl, token)
	if err != nil {
		logrus.Warnf("cannot decorate commits: %s", err)
	}

	commits, err = decorateWithSCMData(repoName, commits, dao)
	if err != nil {
		logrus.Errorf("cannot decorate commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	commits, err = decorateWithEventData(repoName, commits, dao)
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

func commitEvents(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	sha := chi.URLParam(r, "sha")
	repoName := fmt.Sprintf("%s/%s", owner, name)

	ctx := r.Context()
	dao := ctx.Value("store").(*store.Store)

	events, err := dao.EventsForRepoAndSha(repoName, sha)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commitEvents := []*api.CommitEvent{}
	for _, event := range events {
		commitEvents = append(commitEvents, AsCommitEvent(event))
	}

	eventsString, err := json.Marshal(commitEvents)
	if err != nil {
		logrus.Errorf("cannot serialize commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(eventsString)
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

func decorateWithEventData(repoName string, commits []*Commit, dao *store.Store) ([]*Commit, error) {
	hashes := []string{}
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	events, err := dao.LatestEventByRepoAndSha(repoName, hashes)
	if err != nil {
		return nil, err
	}

	eventsByHash := map[string]*model.Event{}
	for _, e := range events {
		eventsByHash[e.SHA] = e
	}

	for _, commit := range commits {
		event := eventsByHash[commit.SHA]
		commit.LastEvent = AsCommitEvent(event)
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
	LastEvent     *api.CommitEvent     `json:"lastEvent,omitempty"`
}

func AsCommitEvent(event *model.Event) *api.CommitEvent {
	if event == nil {
		return nil
	}

	var releaseRequest *dx.ReleaseRequest
	var imageBuildRequest *dx.ImageBuildRequest
	var rollbackRequest *dx.RollbackRequest
	if event.Type == model.ReleaseRequestedEvent {
		var r dx.ReleaseRequest
		err := json.Unmarshal([]byte(event.Blob), &r)
		if err != nil {
			logrus.Warnf("could not unmarshal blob for: %s - %s", event.ID, err)
		}
		releaseRequest = &r
	} else if event.Type == model.ImageBuildRequestedEvent {
		var r dx.ImageBuildRequest
		err := json.Unmarshal([]byte(event.Blob), imageBuildRequest)
		if err != nil {
			logrus.Warnf("could not unmarshal blob for: %s - %s", event.ID, err)
		}
		imageBuildRequest = &r
	} else if event.Type == model.RollbackRequestedEvent {
		var r dx.RollbackRequest
		err := json.Unmarshal([]byte(event.Blob), rollbackRequest)
		if err != nil {
			logrus.Warnf("could not unmarshal blob for: %s - %s", event.ID, err)
		}
		rollbackRequest = &r
	}

	results := []api.CommitEventResult{}
	for _, r := range event.Results {
		var app string
		var env string
		if r.Manifest != nil {
			app = r.Manifest.App
			env = r.Manifest.Env
		}

		results = append(results, api.CommitEventResult{
			App:         app,
			Env:         env,
			Status:      r.Status.String(),
			StatusDesc:  r.StatusDesc,
			GitopsRef:   r.GitopsRef,
			GitopsRepo:  r.GitopsRepo,
			TriggeredBy: r.TriggeredBy,
			Log:         r.Log,
		})
	}

	return &api.CommitEvent{
		ID:                event.ID,
		Created:           event.Created,
		Type:              event.Type,
		Sha:               event.SHA,
		ReleaseRequest:    releaseRequest,
		ImageBuildRequest: imageBuildRequest,
		RollbackRequest:   rollbackRequest,
		Status:            event.Status,
		StatusDesc:        event.StatusDesc,
		Results:           results,
	}
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
