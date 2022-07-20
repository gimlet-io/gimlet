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
	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker/events"
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
	parsedGitopsRepos       map[string]*config.GitopsRepoConfig
	gitopsRepoDeployKeyPath string
	tokenManager            customScm.NonImpersonatedTokenManager
	notificationsManager    notifications.Manager
	eventsProcessed         prometheus.Counter
	repoCache               *nativeGit.GitopsRepoCache
	eventSinkHub            *streaming.EventSinkHub
}

func NewGitopsWorker(
	store *store.Store,
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	tokenManager customScm.NonImpersonatedTokenManager,
	notificationsManager notifications.Manager,
	eventsProcessed prometheus.Counter,
	repoCache *nativeGit.GitopsRepoCache,
	eventSinkHub *streaming.EventSinkHub,
) *GitopsWorker {
	return &GitopsWorker{
		store:                   store,
		gitopsRepo:              gitopsRepo,
		parsedGitopsRepos:       parsedGitopsRepos,
		gitopsRepoDeployKeyPath: gitopsRepoDeployKeyPath,
		notificationsManager:    notificationsManager,
		tokenManager:            tokenManager,
		eventsProcessed:         eventsProcessed,
		repoCache:               repoCache,
		eventSinkHub:            eventSinkHub,
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
				w.parsedGitopsRepos,
				w.gitopsRepoDeployKeyPath,
				w.tokenManager,
				event,
				w.notificationsManager,
				w.repoCache,
				w.eventSinkHub,
			)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func processEvent(
	store *store.Store,
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	tokenManager customScm.NonImpersonatedTokenManager,
	event *model.Event,
	notificationsManager notifications.Manager,
	repoCache *nativeGit.GitopsRepoCache,
	eventSinkHub *streaming.EventSinkHub,
) {
	var token string
	if tokenManager != nil { // only needed for private helm charts
		token, _, _ = tokenManager.Token()
	}

	// process event based on type
	var err error
	var deployEvents []model.Result
	var rollbackEvent *events.RollbackEvent
	var deleteEvents []*events.DeleteEvent
	switch event.Type {
	case model.ArtifactCreatedEvent:
		deployEvents, err = processArtifactEvent(
			gitopsRepo,
			parsedGitopsRepos,
			repoCache,
			gitopsRepoDeployKeyPath,
			token,
			event,
			store,
		)
	case model.ReleaseRequestedEvent:
		deployEvents, err = processReleaseEvent(
			store,
			gitopsRepo,
			parsedGitopsRepos,
			repoCache,
			gitopsRepoDeployKeyPath,
			token,
			event,
		)
	case model.RollbackRequestedEvent:
		rollbackEvent, err = processRollbackEvent(
			gitopsRepo,
			parsedGitopsRepos,
			gitopsRepoDeployKeyPath,
			repoCache,
			event,
		)
		notificationsManager.Broadcast(notifications.MessageFromRollbackEvent(rollbackEvent))
		for _, sha := range rollbackEvent.GitopsRefs {
			setGitopsHashOnEvent(event, sha)
			saveAndBroadcastRollbackEvent(rollbackEvent, sha, event, store, eventSinkHub)
		}
	case model.BranchDeletedEvent:
		deleteEvents, err = processBranchDeletedEvent(
			gitopsRepo,
			parsedGitopsRepos,
			gitopsRepoDeployKeyPath,
			repoCache,
			event,
		)
		for _, deleteEvent := range deleteEvents {
			notificationsManager.Broadcast(notifications.MessageFromDeleteEvent(deleteEvent))
			setGitopsHashOnEvent(event, deleteEvent.GitopsRef)
		}
	}

	// send out notifications based on gitops events
	for _, deployEvent := range deployEvents {
		notificationsManager.Broadcast(notifications.MessageFromGitOpsEvent(deployEvent))
	}

	// record gitops hashes on events
	event.Results = []model.Result{}
	for _, deployEvent := range deployEvents {
		setGitopsHashOnEvent(event, deployEvent.GitopsRef)
		saveAndBroadcastGitopsCommit(deployEvent, event, store, eventSinkHub)
		event.Results = append(event.Results, deployEvent)
	}

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
}

func processBranchDeletedEvent(
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	event *model.Event,
) ([]*events.DeleteEvent, error) {
	var deletedEvents []*events.DeleteEvent
	var branchDeletedEvent events.BranchDeletedEvent
	err := json.Unmarshal([]byte(event.Blob), &branchDeletedEvent)
	if err != nil {
		return nil, fmt.Errorf("cannot parse delete request with id: %s", event.ID)
	}

	for _, env := range branchDeletedEvent.Manifests {
		if env.Cleanup == nil {
			continue
		}

		repoName, _, err := repoInfo(parsedGitopsRepos, env.Env, gitopsRepo)
		if err != nil {
			return nil, fmt.Errorf("cannot find repository to write")
		}

		gitopsEvent := &events.DeleteEvent{
			Env:         env.Env,
			App:         env.Cleanup.AppToCleanup,
			TriggeredBy: "policy",
			Status:      events.Success,
			GitopsRepo:  repoName,

			BranchDeletedEvent: branchDeletedEvent,
		}

		err = env.Cleanup.ResolveVars(map[string]string{
			"BRANCH": branchDeletedEvent.Branch,
		})
		if err != nil {
			gitopsEvent.Status = events.Failure
			gitopsEvent.StatusDesc = err.Error()
			return []*events.DeleteEvent{gitopsEvent}, err
		}
		gitopsEvent.App = env.Cleanup.AppToCleanup // vars are resolved now

		if !cleanupTrigger(branchDeletedEvent.Branch, env.Cleanup) {
			continue
		}

		gitopsEvent, err = cloneTemplateDeleteAndPush(
			gitopsRepo,
			parsedGitopsRepos,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			env.Cleanup,
			env.Env,
			"policy",
			gitopsEvent,
		)
		if gitopsEvent != nil {
			deletedEvents = append(deletedEvents, gitopsEvent)
		}
		if err != nil {
			return deletedEvents, err
		}
	}

	return deletedEvents, err
}

func setGitopsHashOnEvent(event *model.Event, gitopsSha string) {
	if gitopsSha == "" {
		return
	}

	if event.GitopsHashes == nil {
		event.GitopsHashes = []string{}
	}

	event.GitopsHashes = append(event.GitopsHashes, gitopsSha)
}

func processReleaseEvent(
	store *store.Store,
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	gitopsRepoDeployKeyPath string,
	githubChartAccessToken string,
	event *model.Event,
) ([]model.Result, error) {
	var deployEvents []model.Result
	var releaseRequest dx.ReleaseRequest
	err := json.Unmarshal([]byte(event.Blob), &releaseRequest)
	if err != nil {
		return deployEvents, fmt.Errorf("cannot parse release request with id: %s", event.ID)
	}

	artifactEvent, err := store.Artifact(releaseRequest.ArtifactID)
	if err != nil {
		return deployEvents, fmt.Errorf("cannot find artifact with id: %s", event.ArtifactID)
	}
	artifact, err := model.ToArtifact(artifactEvent)
	if err != nil {
		return deployEvents, fmt.Errorf("cannot parse artifact %s", err.Error())
	}

	manifests, err := artifact.CueEnvironmentsToManifests()
	if err != nil {
		return deployEvents, err
	}
	artifact.Environments = append(artifact.Environments, manifests...)

	for _, manifest := range artifact.Environments {
		if manifest.Env != releaseRequest.Env {
			continue
		}

		repoName, _, err := repoInfo(parsedGitopsRepos, manifest.Env, gitopsRepo)
		if err != nil {
			return deployEvents, err
		}

		deployEvent := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: releaseRequest.TriggeredBy,
			Status:      model.Success,
			GitopsRepo:  repoName,
		}

		err = manifest.ResolveVars(artifact.CollectVariables())
		if err != nil {
			deployEvent.Status = model.Failure
			deployEvent.StatusDesc = err.Error()
			deployEvents = append(deployEvents, deployEvent)
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
			parsedGitopsRepos,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			githubChartAccessToken,
			manifest,
			releaseMeta,
		)
		if err != nil {
			deployEvent.Status = model.Failure
			deployEvent.StatusDesc = err.Error()
		}
		deployEvent.GitopsRef = sha
		deployEvents = append(deployEvents, deployEvent)
	}

	return deployEvents, nil
}

func processRollbackEvent(
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoDeployKeyPath string,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	event *model.Event,
) (*events.RollbackEvent, error) {
	var rollbackRequest dx.RollbackRequest
	err := json.Unmarshal([]byte(event.Blob), &rollbackRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot parse release request with id: %s", event.ID)
	}

	repoName, repoPerEnv, err := repoInfo(parsedGitopsRepos, rollbackRequest.Env, gitopsRepo)
	if err != nil {
		return nil, fmt.Errorf("cannot find repository to write")
	}

	rollbackEvent := &events.RollbackEvent{
		RollbackRequest: &rollbackRequest,
		GitopsRepo:      repoName,
	}

	t0 := time.Now().UnixNano()
	repo, repoTmpPath, deployKeyPath, err := gitopsRepoCache.InstanceForWrite(rollbackEvent.GitopsRepo)
	logrus.Infof("Obtaining instance for write took %d", (time.Now().UnixNano()-t0)/1000/1000)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		rollbackEvent.Status = events.Failure
		rollbackEvent.StatusDesc = err.Error()
		return rollbackEvent, err
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
		rollbackEvent.Status = events.Failure
		rollbackEvent.StatusDesc = err.Error()
		return rollbackEvent, err
	}

	hashes, err := shasSince(repo, headSha.Hash().String())
	if err != nil {
		rollbackEvent.Status = events.Failure
		rollbackEvent.StatusDesc = err.Error()
		return rollbackEvent, err
	}

	head, _ := repo.Head()
	err = nativeGit.NativePush(repoTmpPath, deployKeyPath, head.Name().Short())
	if err != nil {
		rollbackEvent.Status = events.Failure
		rollbackEvent.StatusDesc = err.Error()
		return rollbackEvent, err
	}
	gitopsRepoCache.Invalidate(repoName)

	rollbackEvent.GitopsRefs = hashes
	rollbackEvent.Status = events.Success
	return rollbackEvent, nil
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
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	gitopsRepoDeployKeyPath string,
	githubChartAccessToken string,
	event *model.Event,
	dao *store.Store,
) ([]model.Result, error) {
	var deployEvents []model.Result
	artifact, err := model.ToArtifact(event)
	if err != nil {
		return deployEvents, fmt.Errorf("cannot parse artifact %s", err.Error())
	}

	if artifact.HasCleanupPolicy() {
		keepReposWithCleanupPolicyUpToDate(dao, artifact)
	}

	manifests, err := artifact.CueEnvironmentsToManifests()
	if err != nil {
		return deployEvents, err
	}
	artifact.Environments = append(artifact.Environments, manifests...)

	for _, manifest := range artifact.Environments {
		repoName, _, err := repoInfo(parsedGitopsRepos, manifest.Env, gitopsRepo)
		if err != nil {
			return deployEvents, err
		}

		deployEvent := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: "policy",
			Status:      model.Success,
			GitopsRepo:  repoName,
		}

		err = manifest.ResolveVars(artifact.CollectVariables())
		if err != nil {
			deployEvent.Status = model.Failure
			deployEvent.StatusDesc = err.Error()
			deployEvents = append(deployEvents, deployEvent)
			continue
		}

		if !deployTrigger(artifact, manifest.Deploy) {
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
			parsedGitopsRepos,
			gitopsRepoCache,
			gitopsRepoDeployKeyPath,
			githubChartAccessToken,
			manifest,
			releaseMeta,
		)
		if err != nil {
			deployEvent.Status = model.Failure
			deployEvent.StatusDesc = err.Error()
		}
		deployEvent.GitopsRef = sha
		deployEvents = append(deployEvents, deployEvent)
	}

	return deployEvents, nil
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
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	gitopsRepoDeployKeyPath string,
	githubChartAccessToken string,
	manifest *dx.Manifest,
	releaseMeta *dx.Release,
) (string, error) {
	repoName, repoPerEnv, err := repoInfo(parsedGitopsRepos, manifest.Env, gitopsRepo)
	if err != nil {
		return "", err
	}

	repo, repoTmpPath, deployKeyPath, err := gitopsRepoCache.InstanceForWrite(repoName)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		return "", err
	}

	sha, err := gitopsTemplateAndWrite(
		repo,
		manifest,
		releaseMeta,
		githubChartAccessToken,
		repoPerEnv,
	)
	if err != nil {
		return "", err
	}

	if sha != "" { // if there is a change to push
		head, _ := repo.Head()

		operation := func() error {
			return nativeGit.NativePush(repoTmpPath, deployKeyPath, head.Name().Short())
		}
		backoffStrategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
		err := backoff.Retry(operation, backoffStrategy)
		if err != nil {
			return "", err
		}
		gitopsRepoCache.Invalidate(repoName)
	}

	return sha, nil
}

func cloneTemplateDeleteAndPush(
	gitopsRepo string,
	parsedGitopsRepos map[string]*config.GitopsRepoConfig,
	gitopsRepoCache *nativeGit.GitopsRepoCache,
	gitopsRepoDeployKeyPath string,
	cleanupPolicy *dx.Cleanup,
	env string,
	triggeredBy string,
	gitopsEvent *events.DeleteEvent,
) (*events.DeleteEvent, error) {
	repoName, repoPerEnv, err := repoInfo(parsedGitopsRepos, env, gitopsRepo)
	if err != nil {
		gitopsEvent.Status = events.Failure
		gitopsEvent.StatusDesc = err.Error()
		return gitopsEvent, err
	}

	repo, repoTmpPath, deployKeyPath, err := gitopsRepoCache.InstanceForWrite(repoName)
	defer nativeGit.TmpFsCleanup(repoTmpPath)
	if err != nil {
		gitopsEvent.Status = events.Failure
		gitopsEvent.StatusDesc = err.Error()
		return gitopsEvent, err
	}

	path := filepath.Join(env, cleanupPolicy.AppToCleanup)
	if repoPerEnv {
		path = cleanupPolicy.AppToCleanup
	}

	err = nativeGit.DelDir(repo, path)
	if err != nil {
		gitopsEvent.Status = events.Failure
		gitopsEvent.StatusDesc = err.Error()
		return gitopsEvent, err
	}

	empty, err := nativeGit.NothingToCommit(repo)
	if err != nil {
		gitopsEvent.Status = events.Failure
		gitopsEvent.StatusDesc = err.Error()
		return gitopsEvent, err
	}
	if empty {
		return nil, nil
	}

	gitMessage := fmt.Sprintf("[GimletD delete] %s/%s deleted by %s", env, cleanupPolicy.AppToCleanup, triggeredBy)
	sha, err := nativeGit.Commit(repo, gitMessage)

	if sha != "" { // if there is a change to push
		err = nativeGit.Push(repo, deployKeyPath)
		if err != nil {
			gitopsEvent.Status = events.Failure
			gitopsEvent.StatusDesc = err.Error()
			return gitopsEvent, err
		}
		gitopsRepoCache.Invalidate(repoName)

		gitopsEvent.GitopsRef = sha
	}

	return gitopsEvent, nil
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

		if !nativeGit.RollbackCommit(c) {
			commitsToRevert = append(commitsToRevert, c)
		}
		return nil
	})
	if err != nil && err.Error() != "EOF" {
		return err
	}

	for _, commit := range commitsToRevert {
		hasBeenReverted, err := nativeGit.HasBeenReverted(repo, commit, env, app, repoPerEnv)
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
	gitopsHashesString, err := json.Marshal(event.GitopsHashes)
	if err != nil {
		return err
	}
	resultsString, err := json.Marshal(event.Results)
	if err != nil {
		return err
	}
	return store.UpdateEventStatus(event.ID, event.Status, event.StatusDesc, string(gitopsHashesString), string(resultsString))
}

func gitopsTemplateAndWrite(
	repo *git.Repository,
	manifest *dx.Manifest,
	release *dx.Release,
	tokenForChartClone string,
	repoPerEnv bool,
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

func saveAndBroadcastGitopsCommit(deployEvent model.Result, event *model.Event, store *store.Store, eventSinkHub *streaming.EventSinkHub) {
	if deployEvent.GitopsRef == "" {
		return
	}

	gitopsCommitToSave := model.GitopsCommit{
		Sha:        deployEvent.GitopsRef,
		Status:     model.NotReconciled,
		StatusDesc: deployEvent.StatusDesc,
		Created:    event.Created,
		Env:        deployEvent.Manifest.Env,
	}

	eventSinkHub.BroadcastEvent(&gitopsCommitToSave)

	err := store.SaveOrUpdateGitopsCommit(&gitopsCommitToSave)
	if err != nil {
		logrus.Warnf("could not save or update gitops commit: %s", err)
	}
}

func saveAndBroadcastRollbackEvent(rollbackEvent *events.RollbackEvent, sha string, event *model.Event, store *store.Store, eventSinkHub *streaming.EventSinkHub) {
	var rollbackStatus string
	if rollbackEvent.Status == 0 {
		rollbackStatus = model.ReconciliationSucceeded
	} else {
		rollbackStatus = model.ReconciliationFailed
	}
	gitopsCommitToSave := model.GitopsCommit{
		Sha:        sha,
		Status:     rollbackStatus,
		StatusDesc: rollbackEvent.StatusDesc,
		Created:    event.Created,
		Env:        rollbackEvent.RollbackRequest.Env,
	}

	eventSinkHub.BroadcastEvent(&gitopsCommitToSave)

	err := store.SaveOrUpdateGitopsCommit(&gitopsCommitToSave)
	if err != nil {
		logrus.Warnf("could not save or update gitops commit: %s", err)
	}
}

func repoInfo(parsedGitopsRepos map[string]*config.GitopsRepoConfig, env string, defaultGitopsRepo string) (string, bool, error) {
	repoName := defaultGitopsRepo
	repoPerEnv := false

	if repoConfig, ok := parsedGitopsRepos[env]; ok {
		repoName = repoConfig.GitopsRepo
		repoPerEnv = repoConfig.RepoPerEnv
	}

	if repoName == "" {
		return "", false, errors.Errorf("could not find repository for %s environment and GITOPS_REPO did not provide a default repository", env)
	}

	return repoName, repoPerEnv, nil
}
