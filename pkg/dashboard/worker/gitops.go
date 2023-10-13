package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	bootstrap "github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops/sync"
	"github.com/joho/godotenv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type GitopsWorker struct {
	store                   *store.Store
	gitopsRepo              string
	gitopsRepoDeployKeyPath string
	tokenManager            customScm.NonImpersonatedTokenManager
	notificationsManager    notifications.Manager
	eventsProcessed         prometheus.Counter
	repoCache               *nativeGit.RepoCache
	clientHub               *streaming.ClientHub
	perf                    *prometheus.HistogramVec
	gitUser                 *model.User
	gitHost                 string
}

func NewGitopsWorker(
	store *store.Store,
	gitopsRepo string,
	gitopsRepoDeployKeyPath string,
	tokenManager customScm.NonImpersonatedTokenManager,
	notificationsManager notifications.Manager,
	eventsProcessed prometheus.Counter,
	repoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
) *GitopsWorker {
	return &GitopsWorker{
		store:                   store,
		gitopsRepo:              gitopsRepo,
		gitopsRepoDeployKeyPath: gitopsRepoDeployKeyPath,
		notificationsManager:    notificationsManager,
		tokenManager:            tokenManager,
		eventsProcessed:         eventsProcessed,
		repoCache:               repoCache,
		clientHub:               clientHub,
		perf:                    perf,
		gitUser:                 gitUser,
		gitHost:                 gitHost,
	}
}

func (w *GitopsWorker) Run() {
	for {
		events, err := w.store.UnprocessedEvents()
		if err != nil {
			logrus.Errorf("Could not fetch unprocessed events %s", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}

		for _, event := range events {
			w.eventsProcessed.Inc()
			processEvent(w.store,
				w.gitopsRepo,
				w.gitopsRepoDeployKeyPath,
				w.tokenManager,
				event,
				w.notificationsManager,
				w.repoCache,
				w.clientHub,
				w.perf,
				w.gitUser,
				w.gitHost,
			)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func processEvent(
	store *store.Store,
	gitopsRepo string,
	gitopsRepoDeployKeyPath string,
	tokenManager customScm.NonImpersonatedTokenManager,
	event *model.Event,
	notificationsManager notifications.Manager,
	repoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
) {
	var token string
	if tokenManager != nil { // only needed for private helm charts
		token, _, _ = tokenManager.Token()
	}

	// process event based on type
	var err error
	var results []model.Result
	switch event.Type {
	case model.ArtifactCreatedEvent:
		results, err = processArtifactEvent(
			gitopsRepo,
			repoCache,
			gitopsRepoDeployKeyPath,
			token,
			event,
			store,
			perf,
			gitUser,
			gitHost,
		)
	case model.ReleaseRequestedEvent:
		results, err = processReleaseEvent(
			store,
			gitopsRepo,
			repoCache,
			gitopsRepoDeployKeyPath,
			token,
			event,
			perf,
			gitUser,
			gitHost,
		)
	case model.RollbackRequestedEvent:
		results, err = processRollbackEvent(
			gitopsRepo,
			gitopsRepoDeployKeyPath,
			repoCache,
			event,
			token,
			store,
			gitUser,
			gitHost,
		)
	case model.BranchDeletedEvent:
		results, err = processBranchDeletedEvent(
			gitopsRepo,
			gitopsRepoDeployKeyPath,
			repoCache,
			event,
			token,
			store,
		)
	case model.ImageBuildRequestedEvent:
		requestedAt := time.Unix(event.Created, 0)
		anHourAgo := time.Now().Add(-1 * time.Hour)

		if requestedAt.Before(anHourAgo) {
			err = fmt.Errorf("image build timed out")
		} else {
			return
		}
	}

	// associate gitops writes with events
	event.Results = results

	// store event state
	if err != nil {
		logrus.Errorf("error in processing event: %s", err.Error())
		event.Status = model.StatusError
		event.StatusDesc = err.Error()
		err := updateEvent(store, event)
		if err != nil {
			logrus.Warnf("could not update event status %v", err)
		}
	} else {
		event.Status = model.StatusProcessed
		err := updateEvent(store, event)
		if err != nil {
			logrus.Warnf("could not update event status %v", err)
		}
	}

	// broadcast gitops commits to clients
	for _, result := range results {
		var env string
		if event.Type == model.RollbackRequestedEvent {
			env = result.RollbackRequest.Env
		} else {
			env = result.Manifest.Env
		}
		saveAndBroadcastGitopsCommit(result.GitopsRef, env, event, store, clientHub)
	}

	// send out notifications
	if event.Type == model.RollbackRequestedEvent {
		if event.Results != nil {
			m, err := notifications.MessageFromRollbackEvent(*event)
			if err != nil {
				logrus.Warnf("could not convert to notification %v", err)
			}
			notificationsManager.Broadcast(m)
		}
	} else {
		for _, result := range results {
			switch event.Type {
			case model.ArtifactCreatedEvent:
				fallthrough
			case model.ReleaseRequestedEvent:
				notificationsManager.Broadcast(notifications.DeployMessageFromGitOpsResult(result))
			case model.BranchDeletedEvent:
				notificationsManager.Broadcast(notifications.MessageFromDeleteEvent(result))
			}
		}
	}
}

func processBranchDeletedEvent(
	gitopsRepo string,
	gitopsRepoDeployKeyPath string,
	gitopsRepoCache *nativeGit.RepoCache,
	event *model.Event,
	nonImpersonatedToken string,
	store *store.Store,
) ([]model.Result, error) {
	var branchDeletedEvent dx.BranchDeletedEvent
	err := json.Unmarshal([]byte(event.Blob), &branchDeletedEvent)
	if err != nil {
		return nil, fmt.Errorf("cannot parse delete request with id: %s", event.ID)
	}

	results := []model.Result{}
	for _, env := range branchDeletedEvent.Manifests {
		if env.Cleanup == nil {
			continue
		}

		envFromStore, err := store.GetEnvironment(env.Env)
		if err != nil {
			return nil, err
		}

		result := model.Result{
			Manifest:    env,
			TriggeredBy: "policy",
			GitopsRepo:  envFromStore.AppsRepo,
		}

		vars := map[string]string{
			"BRANCH": branchDeletedEvent.Branch,
		}
		envVars, err := loadEnvVars(gitopsRepoCache, envFromStore)
		if err != nil {
			result.Status = model.Failure
			result.StatusDesc = err.Error()
			results = append(results, result)
			continue
		}

		for k, v := range envVars {
			vars[k] = v
		}

		err = env.Cleanup.ResolveVars(vars)
		if err != nil {
			result.Status = model.Failure
			result.StatusDesc = err.Error()
			results = append(results, result)
			continue
		}

		if !cleanupTrigger(branchDeletedEvent.Branch, env.Cleanup) {
			continue
		}

		sha, err := cloneTemplateDeleteAndPush(
			gitopsRepo,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			env.Cleanup,
			env.Env,
			"policy",
			nonImpersonatedToken,
			store,
		)
		if err != nil {
			result.Status = model.Failure
			result.StatusDesc = err.Error()
			results = append(results, result)
			continue
		}

		result.Status = model.Success
		result.GitopsRef = sha
		result.GitopsRepo = gitopsRepo
		results = append(results, result)
	}

	return results, err
}

func processReleaseEvent(
	store *store.Store,
	gitopsRepo string,
	gitopsRepoCache *nativeGit.RepoCache,
	gitopsRepoDeployKeyPath string,
	nonImpersonatedToken string,
	event *model.Event,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
) ([]model.Result, error) {
	var deployResults []model.Result
	var releaseRequest dx.ReleaseRequest
	err := json.Unmarshal([]byte(event.Blob), &releaseRequest)
	if err != nil {
		return deployResults, fmt.Errorf("cannot parse release request with id: %s", event.ID)
	}

	artifactEvent, err := store.Artifact(releaseRequest.ArtifactID)
	if err != nil {
		return deployResults, fmt.Errorf("cannot find artifact with id: %s", event.ArtifactID)
	}
	artifact, err := model.ToArtifact(artifactEvent)
	if err != nil {
		return deployResults, fmt.Errorf("cannot parse artifact %s", err.Error())
	}

	manifests, err := artifact.CueEnvironmentsToManifests()
	if err != nil {
		return deployResults, err
	}
	artifact.Environments = append(artifact.Environments, manifests...)

	for _, manifest := range artifact.Environments {
		if manifest.Env != releaseRequest.Env {
			continue
		}

		if releaseRequest.Tenant != "" &&
			manifest.Tenant.Name != releaseRequest.Tenant {
			continue
		}

		envFromStore, err := store.GetEnvironment(manifest.Env)
		if err != nil {
			return deployResults, err
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: releaseRequest.TriggeredBy,
			Status:      model.Success,
			GitopsRepo:  envFromStore.AppsRepo,
		}

		vars := artifact.CollectVariables()
		vars["APP"] = releaseRequest.App
		envVars, err := loadEnvVars(gitopsRepoCache, envFromStore)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		for k, v := range envVars {
			vars[k] = v
		}

		err = manifest.ResolveVars(vars)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		if releaseRequest.App != "" &&
			manifest.App != releaseRequest.App {
			continue
		}

		releaseMeta := &dx.Release{
			App:         manifest.App,
			Env:         manifest.Env,
			ArtifactID:  artifact.ID,
			Version:     &artifact.Version,
			TriggeredBy: releaseRequest.TriggeredBy,
		}

		sha, err := cloneTemplateWriteAndPush(
			gitopsRepo,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			nonImpersonatedToken,
			manifest,
			releaseMeta,
			perf,
			store,
			gitUser,
			gitHost,
		)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
		} else if sha == "" {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = "No changes made to the gitops state. Maybe this is the current version already?"
		}
		deployResult.GitopsRef = sha
		deployResults = append(deployResults, deployResult)
	}

	return deployResults, nil
}

func processRollbackEvent(
	gitopsRepo string,
	gitopsRepoDeployKeyPath string,
	gitopsRepoCache *nativeGit.RepoCache,
	event *model.Event,
	nonImpersonatedToken string,
	store *store.Store,
	gitUser *model.User,
	gitHost string,
) ([]model.Result, error) {
	var rollbackRequest dx.RollbackRequest
	err := json.Unmarshal([]byte(event.Blob), &rollbackRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot parse release request with id: %s", event.ID)
	}

	envFromStore, err := store.GetEnvironment(rollbackRequest.Env)
	if err != nil {
		return nil, err
	}

	t0 := time.Now().UnixNano()
	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(envFromStore.AppsRepo)
	logrus.Infof("Obtaining instance for write took %d", (time.Now().UnixNano()-t0)/1000/1000)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return nil, err
	}

	headSha, _ := repo.Head()

	err = revertTo(
		rollbackRequest.Env,
		rollbackRequest.App,
		envFromStore.RepoPerEnv,
		repo,
		repoTmpPath,
		rollbackRequest.TargetSHA,
	)
	if err != nil {
		return nil, err
	}

	hashes, err := shasSince(repo, headSha.Hash().String())
	if err != nil {
		return nil, err
	}

	head, _ := repo.Head()
	url := fmt.Sprintf("https://abc123:%s@github.com/%s.git", nonImpersonatedToken, envFromStore.AppsRepo)
	if envFromStore.BuiltIn {
		url = fmt.Sprintf("http://%s:%s@%s/%s", gitUser.Login, gitUser.Secret, gitHost, envFromStore.AppsRepo)
	}
	err = nativeGit.NativePushWithToken(
		url,
		repoTmpPath,
		head.Name().Short(),
	)
	if err != nil {
		logrus.Errorf("could not push to git with native command: %s", err)
		return nil, fmt.Errorf("could not push to git. Check server logs")
	}
	gitopsRepoCache.InvalidateNow(envFromStore.AppsRepo)

	rollbackResults := []model.Result{}

	for _, hash := range hashes {
		rollbackResults = append(rollbackResults, model.Result{
			RollbackRequest: &rollbackRequest,
			TriggeredBy:     rollbackRequest.TriggeredBy,
			Status:          model.Success,
			GitopsRef:       hash,
			GitopsRepo:      envFromStore.AppsRepo,
		})
	}

	return rollbackResults, nil
}

func shasSince(repo *git.Repository, since string) ([]string, error) {
	var hashes []string
	commitWalker, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return hashes, fmt.Errorf("cannot walk commits: %s", err)
	}

	err = commitWalker.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == since {
			return fmt.Errorf("%s", "FOUND")
		}
		hashes = append(hashes, c.Hash.String())
		return nil
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "FOUND" {
		return hashes, fmt.Errorf("cannot walk commits: %s", err)
	}

	return hashes, nil
}

func processArtifactEvent(
	gitopsRepo string,
	gitopsRepoCache *nativeGit.RepoCache,
	gitopsRepoDeployKeyPath string,
	githubChartAccessToken string,
	event *model.Event,
	dao *store.Store,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
) ([]model.Result, error) {
	var deployResults []model.Result
	artifact, err := model.ToArtifact(event)
	if err != nil {
		return deployResults, fmt.Errorf("cannot parse artifact %s", err.Error())
	}

	if artifact.HasCleanupPolicy() {
		keepReposWithCleanupPolicyUpToDate(dao, artifact)
	}

	manifests, err := artifact.CueEnvironmentsToManifests()
	if err != nil {
		return deployResults, err
	}
	artifact.Environments = append(artifact.Environments, manifests...)

	for _, manifest := range artifact.Environments {
		if !deployTrigger(artifact, manifest.Deploy) {
			continue
		}

		envFromStore, err := dao.GetEnvironment(manifest.Env)
		if err != nil {
			return deployResults, err
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: "policy",
			Status:      model.Success,
			GitopsRepo:  envFromStore.AppsRepo,
		}

		vars := artifact.CollectVariables()
		envVars, err := loadEnvVars(gitopsRepoCache, envFromStore)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		for k, v := range envVars {
			vars[k] = v
		}

		err = manifest.ResolveVars(vars)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		releaseMeta := &dx.Release{
			App:         manifest.App,
			Env:         manifest.Env,
			ArtifactID:  artifact.ID,
			Version:     &artifact.Version,
			TriggeredBy: "policy",
		}

		sha, err := cloneTemplateWriteAndPush(
			gitopsRepo,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			githubChartAccessToken,
			manifest,
			releaseMeta,
			perf,
			dao,
			gitUser,
			gitHost,
		)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
		}
		deployResult.GitopsRef = sha
		deployResults = append(deployResults, deployResult)
	}

	return deployResults, nil
}

func loadEnvVars(repoCache *nativeGit.RepoCache, env *model.Environment) (map[string]string, error) {
	repo, err := repoCache.InstanceForRead(env.AppsRepo)
	if err != nil {
		return nil, err
	}

	varsPath := filepath.Join(env.Name, ".gimlet/vars")
	if env.RepoPerEnv {
		varsPath = ".gimlet/vars"
	}

	envVarsString, err := nativeGit.Content(repo, varsPath)
	if err != nil {
		return nil, err
	}

	return godotenv.Unmarshal(envVarsString)
}

func keepReposWithCleanupPolicyUpToDate(dao *store.Store, artifact *dx.Artifact) {
	reposWithCleanupPolicy, err := dao.ReposWithCleanupPolicy()
	if err != nil && err != sql.ErrNoRows {
		logrus.Warnf("could not load repos with cleanup policy: %s", err)
	}

	repoIsNew := true
	for _, r := range reposWithCleanupPolicy {
		if r == artifact.Version.RepositoryName {
			repoIsNew = false
			break
		}
	}
	if repoIsNew {
		reposWithCleanupPolicy = append(reposWithCleanupPolicy, artifact.Version.RepositoryName)
		err = dao.SaveReposWithCleanupPolicy(reposWithCleanupPolicy)
		if err != nil {
			logrus.Warnf("could not update repos with cleanup policy: %s", err)
		}
	}
}

func cloneTemplateWriteAndPush(
	gitopsRepo string,
	gitopsRepoCache *nativeGit.RepoCache,
	gitopsRepoDeployKeyPath string,
	nonImpersonatedToken string,
	manifest *dx.Manifest,
	releaseMeta *dx.Release,
	perf *prometheus.HistogramVec,
	store *store.Store,
	gitUser *model.User,
	gitHost string,
) (string, error) {
	t0 := time.Now()

	envFromStore, err := store.GetEnvironment(manifest.Env)
	if err != nil {
		return "", err
	}

	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(envFromStore.AppsRepo)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return "", err
	}

	var kustomizationManifest *manifestgen.Manifest
	if envFromStore.KustomizationPerApp {
		kustomizationManifest, err = kustomizationTemplate(
			manifest,
			envFromStore.AppsRepo,
			repoTmpPath,
			envFromStore.RepoPerEnv,
		)
		if err != nil {
			return "", err
		}
	}

	sha, err := gitopsTemplateAndWrite(
		repo,
		manifest,
		releaseMeta,
		nonImpersonatedToken,
		envFromStore.RepoPerEnv,
		kustomizationManifest,
	)
	if err != nil {
		return "", err
	}

	if sha != "" { // if there is a change to push
		operation := func() error {
			head, _ := repo.Head()
			url := fmt.Sprintf("https://abc123:%s@github.com/%s.git", nonImpersonatedToken, envFromStore.AppsRepo)
			if envFromStore.BuiltIn {
				url = fmt.Sprintf("http://%s:%s@%s/%s", gitUser.Login, gitUser.Secret, gitHost, envFromStore.AppsRepo)
			}

			return nativeGit.NativePushWithToken(
				url,
				repoTmpPath,
				head.Name().Short(),
			)

		}
		backoffStrategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
		err := backoff.Retry(operation, backoffStrategy)
		if err != nil {
			return "", err
		}
		gitopsRepoCache.InvalidateNow(envFromStore.AppsRepo)
	}

	perf.WithLabelValues("gitops_cloneTemplateWriteAndPush").Observe(float64(time.Since(t0).Seconds()))
	return sha, nil
}

func cloneTemplateDeleteAndPush(
	gitopsRepo string,
	gitopsRepoCache *nativeGit.RepoCache,
	gitopsRepoDeployKeyPath string,
	cleanupPolicy *dx.Cleanup,
	env string,
	triggeredBy string,
	nonImpersonatedToken string,
	store *store.Store,
) (string, error) {
	envFromStore, err := store.GetEnvironment(env)
	if err != nil {
		return "", err
	}

	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(envFromStore.AppsRepo)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return "", err
	}

	path := filepath.Join(env, cleanupPolicy.AppToCleanup)
	if envFromStore.RepoPerEnv {
		path = cleanupPolicy.AppToCleanup
	}

	if envFromStore.KustomizationPerApp {
		kustomizationFilePath := filepath.Join(env, "flux", fmt.Sprintf("kustomization-%s.yaml", cleanupPolicy.AppToCleanup))
		if envFromStore.RepoPerEnv {
			kustomizationFilePath = filepath.Join("flux", fmt.Sprintf("kustomization-%s.yaml", cleanupPolicy.AppToCleanup))
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return "", err
		}
		_, err = worktree.Remove(kustomizationFilePath)
		if err != nil {
			return "", err
		}
	}

	err = nativeGit.DelDir(repo, path)
	if err != nil {
		return "", err
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		return "", err
	}
	if empty {
		return "", nil
	}

	gitMessage := fmt.Sprintf("[GimletD delete] %s/%s deleted by %s", env, cleanupPolicy.AppToCleanup, triggeredBy)
	sha, err := nativeGit.Commit(repo, gitMessage)

	if sha != "" { // if there is a change to push
		head, _ := repo.Head()
		err = nativeGit.NativePushWithToken(
			fmt.Sprintf("https://abc123:%s@github.com/%s.git", nonImpersonatedToken, envFromStore.AppsRepo),
			repoTmpPath,
			head.Name().Short(),
		)
		if err != nil {
			return "", err
		}
		gitopsRepoCache.InvalidateNow(envFromStore.AppsRepo)
	}

	return sha, nil
}

func revertTo(
	env string,
	app string,
	repoPerEnv bool,
	repo *git.Repository,
	repoTmpPath string,
	sha string) error {

	path := filepath.Join(env, app)
	if repoPerEnv {
		path = app
	}

	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return errors.WithMessage(err, "could not walk commits")
	}
	commits = nativeGit.NewCommitDirIterFromIter(path, commits, repo)

	commitsToRevert := []*object.Commit{}
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == sha {
			return fmt.Errorf("EOF")
		}

		if !gitops.RollbackCommit(c) {
			commitsToRevert = append(commitsToRevert, c)
		}
		return nil
	})
	if err != nil && err.Error() != "EOF" {
		return err
	}

	for _, commit := range commitsToRevert {
		hasBeenReverted, err := gitops.HasBeenReverted(repo, commit, env, app, repoPerEnv)
		if !hasBeenReverted {
			logrus.Infof("reverting %s", commit.Hash.String())
			err = nativeGit.NativeRevert(repoTmpPath, commit.Hash.String())
			if err != nil {
				return errors.WithMessage(err, "could not revert")
			}
		}
	}
	return nil
}

func updateEvent(store *store.Store, event *model.Event) error {
	resultsString, err := json.Marshal(event.Results)
	if err != nil {
		return err
	}
	return store.UpdateEventStatus(event.ID, event.Status, event.StatusDesc, string(resultsString))
}

func gitopsTemplateAndWrite(
	repo *git.Repository,
	manifest *dx.Manifest,
	release *dx.Release,
	tokenForChartClone string,
	repoPerEnv bool,
	kustomizationManifest *manifestgen.Manifest,
) (string, error) {
	if strings.HasPrefix(manifest.Chart.Name, "git@") {
		return "", fmt.Errorf("only HTTPS git repo urls supported in GimletD for git based charts")
	}
	if strings.Contains(manifest.Chart.Name, ".git") {
		t0 := time.Now().UnixNano()
		tmpChartDir, err := dx.CloneChartFromRepo(manifest, tokenForChartClone)
		if err != nil {
			return "", fmt.Errorf("cannot fetch chart from git %s", err.Error())
		}
		logrus.Infof("Cloning chart took %d", (time.Now().UnixNano()-t0)/1000/1000)
		manifest.Chart.Name = tmpChartDir
		defer os.RemoveAll(tmpChartDir)
	}

	t0 := time.Now().UnixNano()
	templatedManifests, err := manifest.Render()
	if err != nil {
		return "", fmt.Errorf("cannot run render template %s", err.Error())
	}
	logrus.Infof("Helm template took %d", (time.Now().UnixNano()-t0)/1000/1000)

	files := dx.SplitHelmOutput(map[string]string{"manifest.yaml": templatedManifests})

	if kustomizationManifest != nil {
		files[kustomizationManifest.Path] = kustomizationManifest.Content
	}

	releaseString, err := json.Marshal(release)
	if err != nil {
		return "", fmt.Errorf("cannot marshal release meta data %s", err.Error())
	}

	sha, err := nativeGit.CommitFilesToGit(
		repo,
		files,
		manifest.Env,
		manifest.App,
		repoPerEnv,
		"automated deploy",
		string(releaseString))
	if err != nil {
		return "", fmt.Errorf("cannot write to git: %s", err.Error())
	}

	return sha, nil
}

func deployTrigger(artifactToCheck *dx.Artifact, deployPolicy *dx.Deploy) bool {
	if deployPolicy == nil {
		return false
	}

	if deployPolicy.Branch == "" &&
		deployPolicy.Event == nil &&
		deployPolicy.Tag == "" {
		return false
	}

	if deployPolicy.Branch != "" &&
		(deployPolicy.Event == nil || *deployPolicy.Event != *dx.PushPtr() && *deployPolicy.Event != *dx.PRPtr()) {
		return false
	}

	if deployPolicy.Tag != "" &&
		(deployPolicy.Event == nil || *deployPolicy.Event != *dx.TagPtr()) {
		return false
	}

	if deployPolicy.Tag != "" {
		negate := false
		tag := deployPolicy.Branch
		if strings.HasPrefix(deployPolicy.Tag, "!") {
			negate = true
			tag = deployPolicy.Tag[1:]
		}
		g := glob.MustCompile(deployPolicy.Tag)

		exactMatch := tag == artifactToCheck.Version.Tag
		patternMatch := g.Match(artifactToCheck.Version.Tag)

		match := exactMatch || patternMatch

		if negate && match {
			return false
		}
		if !negate && !match {
			return false
		}
	}

	if deployPolicy.Branch != "" {
		negate := false
		branch := deployPolicy.Branch
		if strings.HasPrefix(deployPolicy.Branch, "!") {
			negate = true
			branch = deployPolicy.Branch[1:]
		}
		g := glob.MustCompile(branch)

		exactMatch := branch == artifactToCheck.Version.Branch
		patternMatch := g.Match(artifactToCheck.Version.Branch)

		match := exactMatch || patternMatch

		if negate && match {
			return false
		}
		if !negate && !match {
			return false
		}
	}

	if deployPolicy.Event != nil {
		if *deployPolicy.Event != artifactToCheck.Version.Event {
			return false
		}
	}

	return true
}

func cleanupTrigger(branch string, cleanupPolicy *dx.Cleanup) bool {
	if cleanupPolicy == nil {
		return false
	}

	if cleanupPolicy.Branch == "" {
		return false
	}

	if cleanupPolicy.AppToCleanup == "" {
		return false
	}

	negate := false
	policyBranch := cleanupPolicy.Branch
	if strings.HasPrefix(cleanupPolicy.Branch, "!") {
		negate = true
		branch = cleanupPolicy.Branch[1:]
	}

	g := glob.MustCompile(policyBranch)

	exactMatch := branch == policyBranch
	patternMatch := g.Match(branch)

	match := exactMatch || patternMatch

	if negate && !match {
		return true
	}
	if !negate && match {
		return true
	}

	return false
}

func saveAndBroadcastGitopsCommit(
	sha string,
	env string,
	event *model.Event,
	store *store.Store,
	clientHub *streaming.ClientHub,
) {
	if sha == "" {
		return
	}

	gitopsCommitToSave := model.GitopsCommit{
		Sha:        sha,
		Status:     model.NotReconciled,
		StatusDesc: "",
		Created:    event.Created,
		Env:        env,
	}

	streaming.BroadcastGitopsCommitEvent(clientHub, gitopsCommitToSave)

	_, err := store.SaveOrUpdateGitopsCommit(&gitopsCommitToSave)
	if err != nil {
		logrus.Warnf("could not save or update gitops commit: %s", err)
	}
}

func kustomizationTemplate(
	manifest *dx.Manifest,
	repoName string,
	repoPath string,
	repoPerEnv bool,
) (*manifestgen.Manifest, error) {
	owner, repository := server.ParseRepo(repoName)
	kustomizationName := uniqueKustomizationName(repoPerEnv, owner, repository, manifest.Env, manifest.Namespace, manifest.App)
	fluxPath := filepath.Join(manifest.Env, "flux")
	if repoPerEnv {
		fluxPath = "flux"
	}

	_, gitopsRepoMetaName := bootstrap.GitopsRepoFileAndMetaNameFromRepo(repoPath, fluxPath)
	sourceName := gitopsRepoMetaName
	if sourceName == "" {
		sourceName = bootstrap.UniqueGitopsRepoName(repoPerEnv, owner, repoName, manifest.Env)
	}

	return sync.GenerateKustomizationForApp(
		manifest.App,
		manifest.Env,
		kustomizationName,
		sourceName,
		repoPerEnv)
}

func uniqueKustomizationName(singleEnv bool, owner string, repoName string, env string, namespace string, appName string) string {
	if len(owner) > 10 {
		owner = owner[:10]
	}
	repoName = strings.TrimPrefix(repoName, "gitops-")

	uniqueName := fmt.Sprintf("%s-%s-%s-%s-%s",
		strings.ToLower(owner),
		strings.ToLower(repoName),
		strings.ToLower(env),
		strings.ToLower(namespace),
		strings.ToLower(appName),
	)
	if singleEnv {
		uniqueName = fmt.Sprintf("%s-%s-%s-%s",
			strings.ToLower(owner),
			strings.ToLower(repoName),
			strings.ToLower(namespace),
			strings.ToLower(appName),
		)
	}
	return uniqueName
}
