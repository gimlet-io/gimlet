package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/sirupsen/logrus"
)

// hook processes webhooks from SCMs
// converts it to go-scm objects
// writes to various tables
// triggers async data fetches
func hook(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config := ctx.Value("config").(*config.Config)
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)
	goScmHelper := genericScm.NewGoScmHelper(dynamicConfig, nil)
	gitRepoCache, _ := ctx.Value("gitRepoCache").(*nativeGit.RepoCache)
	clientHub, _ := r.Context().Value("clientHub").(*streaming.ClientHub)

	// duplicating request body as we exhaust it twice
	buf, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	webhook, err := goScmHelper.Parse(r, func(webhook scm.Webhook) (string, error) {
		return config.WebhookSecret, nil
	})
	if err != nil {
		if dynamicConfig.IsGithub() {
			if r.Header.Get("X-GitHub-Event") == "ping" {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("pong"))
				return
			}
			if r.Header.Get("X-GitHub-Event") == "check_run" { // not handled by go-scm, parsing github actions manually
				dao := ctx.Value("store").(*store.Store)
				tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
				token, _, _ := tokenManager.Token()

				r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
				data, err := ioutil.ReadAll(
					io.LimitReader(r.Body, 10000000),
				)
				if err != nil {
					logrus.Errorf("could not get parse webhook body: %s", err)
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				dst := new(checkRunHook)
				err = json.Unmarshal(data, dst)
				if err != nil {
					logrus.Errorf("could not parse webhook: %s", err)
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				gitService := customScm.NewGitService(dynamicConfig)
				processStatusHook(
					dst.Repository.Owner.Login,
					dst.Repository.Name,
					dst.CheckRun.HeadSHA,
					gitRepoCache,
					gitService,
					token,
					dao,
					clientHub,
				)

				writer.WriteHeader(http.StatusOK)
				return
			}
		}

		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	switch webhook.(type) {
	case *scm.PushHook:
		processPushHook(webhook, gitRepoCache)
	case *scm.TagHook:
		processTagHook(webhook)
	case *scm.StatusHook:
		dao := ctx.Value("store").(*store.Store)
		tokenManager := ctx.Value("tokenManager").(customScm.NonImpersonatedTokenManager)
		token, _, _ := tokenManager.Token()

		owner := webhook.Repository().Namespace
		name := webhook.Repository().Name
		w := webhook.(*scm.StatusHook)

		gitService := customScm.NewGitService(dynamicConfig)
		processStatusHook(owner, name, w.SHA, gitRepoCache, gitService, token, dao, clientHub)
	case *scm.BranchHook:
		processBranchHook(webhook, gitRepoCache)
	}

	writer.WriteHeader(http.StatusOK)
}

func processPushHook(webhook scm.Webhook, repoCache *nativeGit.RepoCache) {
	owner := webhook.Repository().Namespace
	name := webhook.Repository().Name

	repoCache.Invalidate(scm.Join(owner, name))
}

func processTagHook(webhook scm.Webhook) {
}

func processStatusHook(
	owner string,
	name string,
	sha string,
	repoCache *nativeGit.RepoCache,
	gitService customScm.CustomGitService,
	token string,
	dao *store.Store,
	clientHub *streaming.ClientHub,
) {
	repo := scm.Join(owner, name)
	commits, err := gitService.FetchCommits(owner, name, token, []string{sha})
	if err != nil {
		logrus.Errorf("Could not fetch commits for %v, %v", repo, err)
		return
	}

	err = dao.SaveCommits(repo, commits)
	if err != nil {
		logrus.Errorf("Could not store commits for %v, %v", repo, err)
		return
	}
	statusOnCommits := map[string]*model.CombinedStatus{}
	for _, c := range commits {
		statusOnCommits[sha] = &c.Status
	}

	broadcastUpdateCommitStatusEvent(clientHub, owner, name, sha, statusOnCommits[sha], nil)

	if len(statusOnCommits) != 0 {
		err = dao.SaveStatusesOnCommits(repo, statusOnCommits)
		if err != nil {
			logrus.Errorf("Could not store status for %v, %v", repo, err)
			return
		}
	}

	repoCache.Invalidate(scm.Join(owner, name))
}

func broadcastUpdateCommitStatusEvent(
	clientHub *streaming.ClientHub,
	owner string,
	name string,
	sha string,
	commitStatus *model.CombinedStatus,
	deployTargets []*api.DeployTarget,
) {
	jsonString, _ := json.Marshal(streaming.CommitStatusUpdatedEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.CommitStatusUpdatedEventString},
		CommitStatus:   commitStatus,
		Owner:          owner,
		RepoName:       name,
		Sha:            sha,
		DeployTargets:  deployTargets,
	})
	clientHub.Broadcast <- jsonString
}

func processBranchHook(webhook scm.Webhook, repoCache *nativeGit.RepoCache) {
	owner := webhook.Repository().Namespace
	name := webhook.Repository().Name

	repoCache.Invalidate(scm.Join(owner, name))
}

type checkRunHook struct {
	CheckRun struct {
		HeadSHA string `json:"head_sha"`
	} `json:"check_run"`
	Repository struct {
		ID    int64 `json:"id"`
		Owner struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"owner"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}
