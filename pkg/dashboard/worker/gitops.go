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
	"github.com/fluxcd/flux2/pkg/manifestgen"
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
		)
	case model.RollbackRequestedEvent:
		results, err = processRollbackEvent(
			gitopsRepo,
			gitopsRepoDeployKeyPath,
			repoCache,
			event,
			token,
			store,
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

		repoName, _, _, err := gitopsRepoForEnv(store, env.Env)
		if err != nil {
			return nil, err
		}

		result := model.Result{
			Manifest:    env,
			TriggeredBy: "policy",
			GitopsRepo:  repoName,
		}

		err = env.Cleanup.ResolveVars(map[string]string{
			"BRANCH": branchDeletedEvent.Branch,
		})
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
	githubChartAccessToken string,
	event *model.Event,
	perf *prometheus.HistogramVec,
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

		repoName, _, _, err := gitopsRepoForEnv(store, manifest.Env)
		if err != nil {
			return deployResults, err
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: releaseRequest.TriggeredBy,
			Status:      model.Success,
			GitopsRepo:  repoName,
		}

		err = manifest.ResolveVars(artifact.CollectVariables())
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
			githubChartAccessToken,
			manifest,
			releaseMeta,
			perf,
			store,
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
) ([]model.Result, error) {
	var rollbackRequest dx.RollbackRequest
	err := json.Unmarshal([]byte(event.Blob), &rollbackRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot parse release request with id: %s", event.ID)
	}

	repoName, repoPerEnv, _, err := gitopsRepoForEnv(store, rollbackRequest.Env)
	if err != nil {
		return nil, err
	}

	t0 := time.Now().UnixNano()
	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(repoName)
	logrus.Infof("Obtaining instance for write took %d", (time.Now().UnixNano()-t0)/1000/1000)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return nil, err
	}

	headSha, _ := repo.Head()

	err = revertTo(
		rollbackRequest.Env,
		rollbackRequest.App,
		repoPerEnv,
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
	err = nativeGit.NativePushWithToken(repoTmpPath, repoName, nonImpersonatedToken, head.Name().Short())
	if err != nil {
		logrus.Errorf("could not push to git with native command: %s", err)
		return nil, fmt.Errorf("could not push to git. Check server logs")
	}
	gitopsRepoCache.InvalidateNow(repoName)

	rollbackResults := []model.Result{}

	for _, hash := range hashes {
		rollbackResults = append(rollbackResults, model.Result{
			RollbackRequest: &rollbackRequest,
			TriggeredBy:     rollbackRequest.TriggeredBy,
			Status:          model.Success,
			GitopsRef:       hash,
			GitopsRepo:      repoName,
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

		repoName, _, _, err := gitopsRepoForEnv(dao, manifest.Env)
		if err != nil {
			return deployResults, err
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: "policy",
			Status:      model.Success,
			GitopsRepo:  repoName,
		}

		err = manifest.ResolveVars(artifact.CollectVariables())
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
	githubChartAccessToken string,
	manifest *dx.Manifest,
	releaseMeta *dx.Release,
	perf *prometheus.HistogramVec,
	store *store.Store,
) (string, error) {
	t0 := time.Now()

	repoName, repoPerEnv, kustomizationPerApp, err := gitopsRepoForEnv(store, manifest.Env)
	if err != nil {
		return "", err
	}

	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(repoName)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return "", err
	}

	var kustomizationManifest *manifestgen.Manifest
	if kustomizationPerApp {
		kustomizationManifest, err = kustomizationTemplate(
			manifest,
			repoName,
			repoPerEnv,
		)
		if err != nil {
			return "", err
		}
	}

	sha, err := gitopsTemplateAndWrite(
		repo,
		manifest,
		releaseMeta,
		githubChartAccessToken,
		repoPerEnv,
		kustomizationManifest,
	)
	if err != nil {
		return "", err
	}

	if sha != "" { // if there is a change to push
		operation := func() error {
			head, _ := repo.Head()
			return nativeGit.NativePushWithToken(repoTmpPath, repoName, githubChartAccessToken, head.Name().Short())
		}
		backoffStrategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
		err := backoff.Retry(operation, backoffStrategy)
		if err != nil {
			return "", err
		}
		gitopsRepoCache.InvalidateNow(repoName)
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
	repoName, repoPerEnv, kustomizationPerApp, err := gitopsRepoForEnv(store, env)
	if err != nil {
		return "", err
	}

	repo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(repoName)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return "", err
	}

	path := filepath.Join(env, cleanupPolicy.AppToCleanup)
	if repoPerEnv {
		path = cleanupPolicy.AppToCleanup
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
		err = nativeGit.NativePushWithToken(repoTmpPath, repoName, nonImpersonatedToken, head.Name().Short())
		if err != nil {
			return "", err
		}
		gitopsRepoCache.InvalidateNow(repoName)
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

	releaseString, err := json.Marshal(release)
	if err != nil {
		return "", fmt.Errorf("cannot marshal release meta data %s", err.Error())
	}

	sha, err := nativeGit.CommitFilesToGit(
		repo,
		files,
		kustomizationManifest,
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

func gitopsRepoForEnv(db *store.Store, env string) (string, bool, bool, error) {
	envsFromDB, err := db.GetEnvironments()
	if err != nil {
		return "", false, false, fmt.Errorf("cannot get environments from database: %s", err)
	}

	for _, e := range envsFromDB {
		if e.Name == env {
			return e.AppsRepo, e.RepoPerEnv, e.KustomizationPerApp, nil
		}
	}
	return "", false, false, fmt.Errorf("no such environment: %s", env)
}

func kustomizationTemplate(
	manifest *dx.Manifest,
	repoName string,
	repoPerEnv bool,
) (*manifestgen.Manifest, error) {
	owner, repository := server.ParseRepo(repoName)
	kustomizationName := uniqueKustomizationName(repoPerEnv, owner, repository, manifest.Env, manifest.Namespace, manifest.App)
	sourceName := fmt.Sprintf("gitops-repo-%s", bootstrap.UniqueName(repoPerEnv, owner, repository, manifest.Env))
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
