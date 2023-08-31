package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
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
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)

	repo, err := gitRepoCache.InstanceForRead(repoName)
	if err != nil {
		logrus.Errorf("cannot get repo: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if branch == "" {
		branch, err = helper.HeadBranch(repo)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logrus.Errorf("cannot get head branch: %s", err)
			return
		}
	}

	hash := helper.BranchHeadHash(repo, branch)

	if hashString != "head" {
		hash = plumbing.NewHash(hashString)
	}

	commitWalker, err := repo.Log(&git.LogOptions{
		From: hash,
	})
	if err != nil {
		logrus.Errorf("cannot walk commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	limit := 10
	commits := []*Commit{}
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

		return nil
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "LIMIT" {
		logrus.Errorf("cannot walk commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dao := ctx.Value("store").(*store.Store)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	gitServiceImpl := customScm.NewGitService(dynamicConfig)
	tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
	token, _, _ := tokenManager.Token()
	commits, err = decorateCommitsWithSCMData(repoName, commits, dao, gitServiceImpl, token)
	if err != nil {
		logrus.Errorf("cannot decorate commits: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	commits, err = decorateCommitsWithGimletArtifacts(commits, dao, repo, owner, repoName)
	if err != nil {
		logrus.Warnf("cannot get deplyotargets: %s", err)
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

func decorateCommitsWithSCMData(
	repo string,
	commits []*Commit,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
) ([]*Commit, error) {
	return decorateCommitsWithSCMDataWithRetry(
		repo,
		commits,
		dao,
		gitServiceImpl,
		token,
		false,
	)
}

func decorateCommitsWithSCMDataWithRetry(
	repo string,
	commits []*Commit,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
	isRetry bool,
) ([]*Commit, error) {
	var hashes []string
	for _, commit := range commits {
		hashes = append(hashes, commit.SHA)
	}

	dbCommits, err := dao.CommitsByRepoAndSHA(repo, hashes)
	if err != nil {
		return nil, fmt.Errorf("cannot get commits from db %s", err)
	}

	dbCommitsByHash := map[string]*model.Commit{}
	for _, dbCommit := range dbCommits {
		dbCommitsByHash[dbCommit.SHA] = dbCommit
	}

	var decoratedCommits []*Commit
	var hashesToFetch []string
	for _, commit := range commits {
		if dbCommit, ok := dbCommitsByHash[commit.SHA]; ok {
			commit.URL = dbCommit.URL
			commit.Author = dbCommit.Author
			commit.AuthorPic = dbCommit.AuthorPic
			commit.Tags = dbCommit.Tags
			commit.Status = dbCommit.Status
		} else {
			hashesToFetch = append(hashesToFetch, commit.SHA)
		}

		decoratedCommits = append(decoratedCommits, commit)
	}

	// if not all commit was decorated, fetch them and try decorating again
	if len(hashesToFetch) > 0 && !isRetry {
		logrus.Infof("Fetching scm data for %s", strings.Join(hashesToFetch, ","))
		// fetch remote commit info, then try to decorate again
		owner, name := scm.Split(repo)
		fetchCommits(owner, name, gitServiceImpl, token, dao, hashesToFetch)
		return decorateCommitsWithSCMDataWithRetry(
			repo,
			commits,
			dao,
			gitServiceImpl,
			token,
			true,
		)
	}

	return decoratedCommits, nil
}

func decorateDeploymentWithSCMData(
	repo string,
	deployment *api.Deployment,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
) (*api.Deployment, error) {
	return decorateDeploymentWithSCMDataWithRetry(
		repo,
		deployment,
		dao,
		gitServiceImpl,
		token,
		false,
	)
}

func decorateDeploymentWithSCMDataWithRetry(
	repo string,
	deployment *api.Deployment,
	dao *store.Store,
	gitServiceImpl customScm.CustomGitService,
	token string,
	isRetry bool,
) (*api.Deployment, error) {
	dbCommits, err := dao.CommitsByRepoAndSHA(repo, []string{deployment.SHA})
	if err != nil {
		return nil, fmt.Errorf("cannot get commits from db %s", err)
	}

	if len(dbCommits) > 0 {
		deployment.CommitMessage = dbCommits[0].Message
	} else {
		if isRetry { // we only retry once
			return deployment, nil
		}
		owner, name := scm.Split(repo)

		// fetch remote commit info, then try to decorate again
		fetchCommits(owner, name, gitServiceImpl, token, dao, []string{deployment.SHA})
		return decorateDeploymentWithSCMDataWithRetry(
			repo,
			deployment,
			dao,
			gitServiceImpl,
			token,
			true,
		)
	}

	return deployment, nil
}

// commits come from go-scm based live git traversal
// we decorate that commit data with data from SCMs with the
// following fields: url, author, author_pic, message, created, tags, status
// URL, Author, AuthorPic, Tags, Status
func fetchCommits(
	owner string,
	repo string,
	gitService customScm.CustomGitService,
	token string,
	store *store.Store,
	hashesToFetch []string,
) {
	if len(hashesToFetch) == 0 {
		return
	}

	commits, err := gitService.FetchCommits(owner, repo, token, hashesToFetch)
	if err != nil {
		logrus.Errorf("Could not fetch commits for %v, %v", repo, err)
		return
	}

	err = store.SaveCommits(scm.Join(owner, repo), commits)
	if err != nil {
		logrus.Errorf("Could not store commits for %v, %v", repo, err)
		return
	}
	statusOnCommits := map[string]*model.CombinedStatus{}
	for _, c := range commits {
		statusOnCommits[c.SHA] = &c.Status
	}

	if len(statusOnCommits) != 0 {
		err = store.SaveStatusesOnCommits(scm.Join(owner, repo), statusOnCommits)
		if err != nil {
			logrus.Errorf("Could not store status for %v, %v", repo, err)
			return
		}
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
