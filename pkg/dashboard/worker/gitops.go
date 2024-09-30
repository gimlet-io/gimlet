package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/cenkalti/backoff/v4"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet/pkg/dashboard/imageBuild"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet/pkg/git/gogit"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	bootstrap "github.com/gimlet-io/gimlet/pkg/gitops"
	"github.com/gimlet-io/gimlet/pkg/gitops/sync"
	"github.com/joho/godotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type GitopsWorker struct {
	store                *store.Store
	tokenManager         customScm.NonImpersonatedTokenManager
	notificationsManager notifications.Manager
	eventsProcessed      prometheus.Counter
	repoCache            *nativeGit.RepoCache
	clientHub            *streaming.ClientHub
	perf                 *prometheus.HistogramVec
	gitUser              *model.User
	gitHost              string
	gitopsQueue          chan int
	agentHub             *streaming.AgentHub
	dynamicConfig        *dynamicconfig.DynamicConfig
}

func NewGitopsWorker(
	store *store.Store,
	tokenManager customScm.NonImpersonatedTokenManager,
	notificationsManager notifications.Manager,
	eventsProcessed prometheus.Counter,
	repoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
	agentHub *streaming.AgentHub,
	dynamicConfig *dynamicconfig.DynamicConfig,
) *GitopsWorker {

	return &GitopsWorker{
		store:                store,
		notificationsManager: notificationsManager,
		tokenManager:         tokenManager,
		eventsProcessed:      eventsProcessed,
		repoCache:            repoCache,
		clientHub:            clientHub,
		perf:                 perf,
		gitUser:              gitUser,
		gitHost:              gitHost,
		gitopsQueue:          make(chan int, 1000),
		agentHub:             agentHub,
		dynamicConfig:        dynamicConfig,
	}
}

func (w *GitopsWorker) Run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer func() {
		ticker.Stop()
	}()

	w.store.SubscribeToEventCreated(func(event *model.Event) {
		w.gitopsQueue <- 1
	})

	w.gitopsQueue <- 1 // start with processing
	logrus.Info("Gitops worker started")
	for {
		select {
		case _, ok := <-w.gitopsQueue:
			if !ok {
				logrus.Info("Gitops worker stopped")
				return
			}
		case <-ticker.C:
		}

		events, err := w.store.UnprocessedEvents()
		if err != nil {
			logrus.Errorf("Could not fetch unprocessed events %s", err.Error())
		}
		for _, event := range events {
			w.eventsProcessed.Inc()
			processEvent(w.store,
				w.tokenManager,
				event,
				w.notificationsManager,
				w.repoCache,
				w.clientHub,
				w.perf,
				w.gitUser,
				w.gitHost,
				w.agentHub,
				w.dynamicConfig,
			)
		}
	}
}

func processEvent(
	store *store.Store,
	tokenManager customScm.NonImpersonatedTokenManager,
	event *model.Event,
	notificationsManager notifications.Manager,
	repoCache *nativeGit.RepoCache,
	clientHub *streaming.ClientHub,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
	agentHub *streaming.AgentHub,
	dynamicConfig *dynamicconfig.DynamicConfig,
) {
	var token string
	if tokenManager != nil { // only needed for private helm charts
		token, _, _ = tokenManager.Token()
	}

	envConfigs, configLoadError := envConfigs(store, repoCache)
	if configLoadError != nil {
		logrus.Warnf("Could not load envConfigs - preview ingresses may get a wrong url")
		envConfigs = map[string]*dx.StackConfig{}
	}

	// process event based on type
	var err error
	var results []model.Result
	switch event.Type {
	case model.ArtifactCreatedEvent:
		results, err = processArtifactEvent(
			repoCache,
			token,
			event,
			store,
			perf,
			gitUser,
			gitHost,
			agentHub,
			envConfigs,
		)
	case model.ReleaseRequestedEvent:
		results, err = processReleaseEvent(
			store,
			repoCache,
			token,
			event,
			perf,
			gitUser,
			gitHost,
			envConfigs,
		)
	case model.RollbackRequestedEvent:
		results, err = processRollbackEvent(
			repoCache,
			event,
			token,
			store,
			gitUser,
			gitHost,
		)
	case model.BranchDeletedEvent:
		results, err = processBranchDeletedEvent(
			repoCache,
			event,
			token,
			store,
			envConfigs,
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

	// comment on github PRs
	if event.Type == model.ArtifactCreatedEvent ||
		event.Type == model.ReleaseRequestedEvent {
		for _, result := range results {
			if result.Manifest.Preview == nil || !*result.Manifest.Preview {
				break
			}
			commentOnPR(result, dynamicConfig, token)
		}
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

func commentOnPR(result model.Result, dynamicConfig *dynamicconfig.DynamicConfig, token string) {
	vars := result.Artifact.CollectVariables()
	gitRepo := vars["REPO"]
	branch := vars["BRANCH"]
	gitSha := vars["SHA"]

	if result.Status == model.Failure {
		err := comment(dynamicConfig, token, gitRepo, branch, fmt.Sprintf(customScm.BodyFailed, result.Manifest.App, gitSha))
		if err != nil {
			logrus.Warnf("could not update comment %v", err)
		}
	} else {
		var hostString string
		if ingress, ok := result.Manifest.Values["ingress"]; ok {
			ingressMap := ingress.(map[string]interface{})
			if host, ok := ingressMap["host"]; ok {
				hostString = host.(string)
			}
		}
		err := comment(dynamicConfig, token, gitRepo, branch, fmt.Sprintf(customScm.BodyReady, result.Manifest.App, gitSha, hostString, hostString))
		if err != nil {
			logrus.Warnf("could not update comment %v", err)
		}
	}
}

func processBranchDeletedEvent(
	gitopsRepoCache *nativeGit.RepoCache,
	event *model.Event,
	nonImpersonatedToken string,
	store *store.Store,
	envConfigs map[string]*dx.StackConfig,
) ([]model.Result, error) {
	var branchDeletedEvent dx.BranchDeletedEvent
	err := json.Unmarshal([]byte(event.Blob), &branchDeletedEvent)
	if err != nil {
		return nil, fmt.Errorf("cannot parse delete request with id: %s", event.ID)
	}

	results := []model.Result{}
	for _, manifest := range branchDeletedEvent.Manifests {
		manifest.PrepPreview(ingressHost(envConfigs[manifest.Env]))
		if manifest.Cleanup == nil {
			continue
		}

		envFromStore, err := store.GetEnvironment(manifest.Env)
		if err != nil {
			return nil, err
		}

		result := model.Result{
			Manifest:    manifest,
			TriggeredBy: "policy",
			GitopsRepo:  envFromStore.AppsRepo,
		}

		err = manifest.Cleanup.ResolveVars(map[string]string{
			"BRANCH": branchDeletedEvent.Branch,
		})
		if err != nil {
			result.Status = model.Failure
			result.StatusDesc = err.Error()
			results = append(results, result)
			continue
		}

		if !cleanupTrigger(branchDeletedEvent.Branch, manifest.Cleanup) {
			continue
		}

		sha, err := cloneTemplateDeleteAndPush(
			gitopsRepoCache,
			manifest.Cleanup,
			manifest.Env,
			"cleanup policy",
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
		result.GitopsRepo = envFromStore.AppsRepo
		results = append(results, result)
	}

	return results, err
}

func processReleaseEvent(
	store *store.Store,
	gitopsRepoCache *nativeGit.RepoCache,
	nonImpersonatedToken string,
	event *model.Event,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
	envConfigs map[string]*dx.StackConfig,
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

	var repoVars map[string]string
	err = gitopsRepoCache.PerformAction(artifact.Version.RepositoryName, func(repo *git.Repository) error {
		var innerErr error
		repoVars, innerErr = loadVars(repo, ".gimlet/vars")
		return innerErr
	})
	if err != nil {
		return deployResults, fmt.Errorf("cannot load vars %s", err.Error())
	}

	for _, manifest := range artifact.Environments {
		if manifest.Env != releaseRequest.Env {
			continue
		}

		envFromStore, err := store.GetEnvironment(manifest.Env)
		if err != nil {
			return deployResults, fmt.Errorf("no such env: %s", manifest.Env)
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: releaseRequest.TriggeredBy,
			Status:      model.Success,
			GitopsRepo:  envFromStore.AppsRepo,
		}

		appsRepo, repoTmpPath, err := gitopsRepoCache.InstanceForWrite(envFromStore.AppsRepo)
		defer gogit.TmpFsCleanup(repoTmpPath)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		varsPath := filepath.Join(envFromStore.Name, ".gimlet/vars")
		if envFromStore.RepoPerEnv {
			varsPath = ".gimlet/vars"
		}

		envVars, err := loadVars(appsRepo, varsPath)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		vars := artifact.CollectVariables()
		vars["APP"] = releaseRequest.App
		for k, v := range envVars {
			vars[k] = v
		}

		manifest.PrepPreview(ingressHost(envConfigs[manifest.Env]))
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
			appsRepo,
			repoTmpPath,
			gitopsRepoCache,
			nonImpersonatedToken,
			manifest,
			releaseMeta,
			perf,
			store,
			repoVars,
			envVars,
			gitUser,
			gitHost,
			envConfigs[manifest.Env],
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

		event.Results = deployResults
		err = updateEvent(store, event)
		if err != nil {
			logrus.Warnf("could not update event status %v", err)
		}
	}

	return deployResults, nil
}

func processRollbackEvent(
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
	defer gogit.TmpFsCleanup(repoTmpPath)
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
		url = fmt.Sprintf("http://%s:%s@%s/%s", gitUser.Login, gitUser.Token, gitHost, envFromStore.AppsRepo)
	}
	err = gogit.NativePushWithToken(
		url,
		repoTmpPath,
		head.Name().Short(),
	)
	if err != nil {
		logrus.Errorf("could not push to git with native command: %s", err)
		return nil, fmt.Errorf("could not push to git. Check server logs")
	}
	gitopsRepoCache.Invalidate(envFromStore.AppsRepo)

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
	gitRepoCache *nativeGit.RepoCache,
	githubChartAccessToken string,
	event *model.Event,
	dao *store.Store,
	perf *prometheus.HistogramVec,
	gitUser *model.User,
	gitHost string,
	agentHub *streaming.AgentHub,
	envConfigs map[string]*dx.StackConfig,
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

	var repoVars map[string]string
	err = gitRepoCache.PerformAction(artifact.Version.RepositoryName, func(repo *git.Repository) error {
		var innerErr error
		repoVars, innerErr = loadVars(repo, ".gimlet/vars")
		return innerErr
	})
	if err != nil {
		return deployResults, fmt.Errorf("cannot load vars %s", err.Error())
	}

	for _, manifest := range artifact.Environments {
		manifest.PrepPreview(ingressHost(envConfigs[manifest.Env]))
		if !deployTrigger(artifact, manifest.Deploy) {
			continue
		}
		if manifest.Cleanup != nil {
			keepReposWithCleanupPolicyUpToDate(dao, artifact)
		}

		deployResult := model.Result{
			Manifest:    manifest,
			Artifact:    artifact,
			TriggeredBy: "policy",
			Status:      model.Success,
		}

		envFromStore, err := dao.GetEnvironment(manifest.Env)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = fmt.Sprintf("could not load environment: %s", manifest.Env)
			deployResults = append(deployResults, deployResult)
			continue
		}

		appsRepo, repoTmpPath, err := gitRepoCache.InstanceForWrite(envFromStore.AppsRepo)
		defer gogit.TmpFsCleanup(repoTmpPath)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		varsPath := filepath.Join(envFromStore.Name, ".gimlet/vars")
		if envFromStore.RepoPerEnv {
			varsPath = ".gimlet/vars"
		}

		envVars, err := loadVars(appsRepo, varsPath)
		if err != nil {
			deployResult.Status = model.Failure
			deployResult.StatusDesc = err.Error()
			deployResults = append(deployResults, deployResult)
			continue
		}

		vars := artifact.CollectVariables()
		vars["APP"] = manifest.App
		for k, v := range envVars {
			vars[k] = v
		}

		strategy := gitops.ExtractImageStrategy(manifest)
		if strategy == "buildpacks" || strategy == "dockerfile" { // image build
			imageRepository, imageTag, dockerfile, registry := gitops.ExtractImageRepoTagDockerfileAndRegistry(manifest, vars)
			// Image push happens inside the cluster, pull is handled by the kubelet that doesn't speak cluster local addresses
			imageRepository = strings.ReplaceAll(imageRepository, "127.0.0.1:32447", "registry.infrastructure.svc.cluster.local:5000")
			imageBuildRequest := &dx.ImageBuildRequest{
				Env:         manifest.Env,
				App:         manifest.App,
				Sha:         artifact.Version.SHA,
				ArtifactID:  artifact.ID,
				TriggeredBy: "policy",
				Image:       imageRepository,
				Tag:         imageTag,
				Dockerfile:  dockerfile,
				Strategy:    strategy,
				Registry:    registry,
			}
			imageBuildEvent, err := imageBuild.TriggerImagebuild(gitRepoCache, agentHub, dao, artifact, imageBuildRequest)
			if err != nil {
				deployResult.Status = model.Failure
				deployResult.StatusDesc = err.Error()
			}
			deployResult.TriggeredImageBuildRequestID = imageBuildEvent.ID
			deployResults = append(deployResults, deployResult)
		} else { // release
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
				appsRepo,
				repoTmpPath,
				gitRepoCache,
				githubChartAccessToken,
				manifest,
				releaseMeta,
				perf,
				dao,
				repoVars,
				envVars,
				gitUser,
				gitHost,
				envConfigs[manifest.Env],
			)
			if err != nil {
				deployResult.Status = model.Failure
				deployResult.StatusDesc = err.Error()
			}
			deployResult.GitopsRepo = envFromStore.AppsRepo
			deployResult.GitopsRef = sha
			deployResults = append(deployResults, deployResult)
		}
	}

	return deployResults, nil
}

func comment(
	dynamicConfig *dynamicconfig.DynamicConfig,
	token string,
	gitRepo, branch string,
	commentBody string,
) error {
	gitSvc := customScm.NewGitService(dynamicConfig)

	goScm := genericScm.NewGoScmHelper(dynamicConfig, nil)
	openpullrequests, err := goScm.ListOpenPRs(token, gitRepo)
	if err != nil {
		return fmt.Errorf("cannot list open pullrequests: %s", err)
	}

	var pullNumber int
	for _, pr := range openpullrequests {
		if pr.Source == branch {
			pullNumber = pr.Number
		}
	}
	if pullNumber == 0 {
		return nil
	}

	comments, err := gitSvc.Comments(token, gitRepo, pullNumber)
	if err != nil {
		return fmt.Errorf("cannot list comments: %s", err)
	}

	for _, c := range comments {
		if strings.Contains(*c.Body, "Deploy Preview for") {
			return gitSvc.UpdateComment(
				token,
				gitRepo,
				*c.ID,
				commentBody,
			)
		}
	}

	return gitSvc.CreateComment(
		token,
		gitRepo,
		pullNumber,
		commentBody,
	)
}

func loadVars(repo *git.Repository, varsPath string) (map[string]string, error) {
	envVarsString, err := gogit.Content(repo, varsPath)
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
	repo *git.Repository,
	repoTmpPath string,
	gitopsRepoCache *nativeGit.RepoCache,
	nonImpersonatedToken string,
	manifest *dx.Manifest,
	releaseMeta *dx.Release,
	perf *prometheus.HistogramVec,
	store *store.Store,
	repoVars map[string]string,
	envVars map[string]string,
	gitUser *model.User,
	gitHost string,
	stackConfig *dx.StackConfig,
) (string, error) {
	t0 := time.Now()

	environment, err := store.GetEnvironment(manifest.Env)
	if err != nil {
		return "", err
	}

	var kustomizationManifest *manifestgen.Manifest
	if environment.KustomizationPerApp {
		kustomizationManifest, err = kustomizationTemplate(
			manifest,
			environment.AppsRepo,
			repoTmpPath,
			environment.RepoPerEnv,
		)
		if err != nil {
			return "", err
		}
	}

	imagepullSecretManifest, err := imagepullSecretTemplate(
		manifest,
		stackConfig,
		environment.RepoPerEnv,
	)
	if err != nil {
		return "", err
	}

	owner, repository := server.ParseRepo(releaseMeta.Version.RepositoryName)
	perRepoConfigMapName := fmt.Sprintf("%s-%s", strings.ToLower(owner), strings.ToLower(repository))
	perRepoConfigMapManifest, err := sync.GenerateConfigMap(perRepoConfigMapName, manifest.Namespace, repoVars)
	if err != nil {
		return "", err
	}

	perEnvConfigMapManifest, err := sync.GenerateConfigMap(strings.ToLower(environment.Name), manifest.Namespace, envVars)
	if err != nil {
		return "", err
	}

	sha, err := gitopsTemplateAndWrite(
		repo,
		manifest,
		releaseMeta,
		nonImpersonatedToken,
		environment.RepoPerEnv,
		kustomizationManifest,
		imagepullSecretManifest,
		perRepoConfigMapManifest,
		perEnvConfigMapManifest,
	)
	if err != nil {
		return "", err
	}

	if sha != "" { // if there is a change to push
		operation := func() error {
			head, _ := repo.Head()
			url := fmt.Sprintf("https://abc123:%s@github.com/%s.git", nonImpersonatedToken, environment.AppsRepo)
			if environment.BuiltIn {
				url = fmt.Sprintf("http://%s:%s@%s/%s", gitUser.Login, gitUser.Token, gitHost, environment.AppsRepo)
			}

			return gogit.NativePushWithToken(
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
		gitopsRepoCache.Invalidate(environment.AppsRepo)
	}

	perf.WithLabelValues("gitops_cloneTemplateWriteAndPush").Observe(float64(time.Since(t0).Seconds()))
	return sha, nil
}

func cloneTemplateDeleteAndPush(
	gitopsRepoCache *nativeGit.RepoCache,
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
	defer gogit.TmpFsCleanup(repoTmpPath)
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
		err := gogit.DelFile(repo, kustomizationFilePath)
		if err != nil {
			return "", err
		}
	}

	err = gogit.DelDir(repo, path)
	if err != nil {
		return "", err
	}

	empty, err := gogit.NothingToCommit(repo)
	if err != nil {
		return "", err
	}
	if empty {
		return "", nil
	}

	gitMessage := fmt.Sprintf("[Gimlet] %s/%s deleted by %s", env, cleanupPolicy.AppToCleanup, triggeredBy)
	sha, err := gogit.Commit(repo, gitMessage)

	if sha != "" { // if there is a change to push
		head, _ := repo.Head()
		err = gogit.NativePushWithToken(
			fmt.Sprintf("https://abc123:%s@github.com/%s.git", nonImpersonatedToken, envFromStore.AppsRepo),
			repoTmpPath,
			head.Name().Short(),
		)
		if err != nil {
			return "", err
		}
		gitopsRepoCache.Invalidate(envFromStore.AppsRepo)
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
	commits = gogit.NewCommitDirIterFromIter(path, commits, repo)

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
			err = gogit.NativeRevert(repoTmpPath, commit.Hash.String())
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
	imagepullsecretManifest *manifestgen.Manifest,
	perRepoConfigMapManifest *manifestgen.Manifest,
	perEnvConfigMapManifest *manifestgen.Manifest,
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

	helmGeneratedFiles := dx.SplitHelmOutput(map[string]string{"manifest.yaml": templatedManifests})

	envReleaseJsonPath := manifest.Env
	appFolderPath := filepath.Join(manifest.Env, manifest.App)
	if repoPerEnv {
		appFolderPath = manifest.App
		envReleaseJsonPath = ""
	}

	files := map[string]string{}
	for fileName, content := range helmGeneratedFiles {
		files[filepath.Join(appFolderPath, fileName)] = content
	}

	if kustomizationManifest != nil {
		files[kustomizationManifest.Path] = kustomizationManifest.Content
	}

	if imagepullsecretManifest != nil {
		files[imagepullsecretManifest.Path] = imagepullsecretManifest.Content
	}

	if perRepoConfigMapManifest != nil {
		files[perRepoConfigMapManifest.Path] = perRepoConfigMapManifest.Content
	}

	if perEnvConfigMapManifest != nil {
		files[perEnvConfigMapManifest.Path] = perEnvConfigMapManifest.Content
	}

	releaseString, err := json.Marshal(release)
	if err != nil {
		return "", fmt.Errorf("cannot marshal release meta data %s", err.Error())
	}
	files[filepath.Join(appFolderPath, "release.json")] = string(releaseString)
	files[filepath.Join(envReleaseJsonPath, "release.json")] = string(releaseString)

	sha, err := gogit.CommitFilesToGit(
		repo,
		files,
		[]string{appFolderPath},
		fmt.Sprintf("[Gimlet] %s/%s automated deploy", manifest.Env, manifest.App),
	)
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
		deployPolicy.Tag == "" &&
		len(deployPolicy.CommitMessagePatterns) == 0 {
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

	if len(deployPolicy.CommitMessagePatterns) != 0 {
		if !commitMessagePatternMatch(deployPolicy.CommitMessagePatterns, artifactToCheck.Version.Message) {
			return false
		}
	}

	return true
}

func commitMessagePatternMatch(patterns []string, commitMessage string) bool {
	deployAllPattern := glob.MustCompile(escapeSquareBracketChars("*[DEPLOY: ALL]*"))
	if deployAllPattern.Match(commitMessage) {
		return true
	}

	for _, pattern := range patterns {
		g := glob.MustCompile(escapeSquareBracketChars(pattern))
		if g.Match(commitMessage) {
			return true
		}
	}

	return false
}

func escapeSquareBracketChars(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "[", "\\[")
	pattern = strings.ReplaceAll(pattern, "]", "\\]")
	return pattern
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

	_, gitopsRepoMetaName := bootstrap.GitopsRepoFileAndMetaNameFromRepo(repoPath, fluxPath, "")
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

func imagepullSecretTemplate(
	manifest *dx.Manifest,
	stackConfig *dx.StackConfig,
	repoPerEnv bool,
) (*manifestgen.Manifest, error) {
	var registryString string
	if image, ok := manifest.Values["image"]; ok {
		imageMap := image.(map[string]interface{})
		if registry, ok := imageMap["registry"]; ok {
			registryString = registry.(string)
		}
	}
	if registryString == "" {
		return nil, nil
	}
	registry, ok := stackConfig.Config[registryString]
	if !ok {
		return nil, nil
	}
	registryMap := registry.(map[string]interface{})

	credentials, ok := registryMap["credentials"]
	if !ok {
		return nil, nil
	}
	credentialsMap := credentials.(map[string]interface{})

	var encryptedConfigString string
	if encryptedConfig, ok := credentialsMap["encryptedDockerconfigjson"]; ok {
		encryptedConfigString = encryptedConfig.(string)
	}

	return generateImagePullSecret(
		manifest.Env,
		manifest.App,
		manifest.Namespace,
		strings.ToLower(registryString),
		encryptedConfigString,
		repoPerEnv,
	)
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

func generateImagePullSecret(
	env, app, namespace, registry string,
	encryptedDockerconfigjson string,
	singleEnv bool,
) (*manifestgen.Manifest, error) {
	secretPath := filepath.Join(env, app)
	if singleEnv {
		secretPath = app
	}

	secretName := fmt.Sprintf("%s-%s-pullsecret", app, registry)
	secret := ssv1alpha1.SealedSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SealedSecret",
			APIVersion: "bitnami.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Annotations: map[string]string{
				"sealedsecrets.bitnami.com/cluster-wide": "true",
			},
		},
		Spec: ssv1alpha1.SealedSecretSpec{
			EncryptedData: ssv1alpha1.SealedSecretEncryptedData{
				".dockerconfigjson": encryptedDockerconfigjson,
			},
			Template: ssv1alpha1.SecretTemplateSpec{
				Type: "kubernetes.io/dockerconfigjson",
			},
		},
	}

	secretData, err := yaml.Marshal(secret)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(secretPath, fmt.Sprintf("imagepullsecret-%s.yaml", app)),
		Content: fmt.Sprintf("---\n%s", string(secretData)),
	}, nil
}

func ingressHost(envConfig *dx.StackConfig) string {
	if envConfig == nil {
		return ""
	}

	if val, ok := envConfig.Config["existingIngress"]; ok {
		existingIngress := val.(map[string]interface{})
		if val, ok := existingIngress["host"]; ok {
			return val.(string)
		} else {
			return ""
		}
	} else if val, ok := envConfig.Config["nginx"]; ok {
		nginx := val.(map[string]interface{})
		if val, ok := nginx["host"]; ok {
			return val.(string)
		} else {
			return ""
		}
	} else {
		return ""
	}
}

func envConfigs(
	dao *store.Store,
	gitRepoCache *nativeGit.RepoCache,
) (map[string]*dx.StackConfig, error) {
	environments, err := dao.GetEnvironments()
	if err != nil {
		return nil, err
	}

	envConfigs := map[string]*dx.StackConfig{}
	for _, env := range environments {
		stackYamlPath := "stack.yaml"
		if !env.RepoPerEnv {
			stackYamlPath = filepath.Join(env.Name, "stack.yaml")
		}

		stackConfig, err := server.StackConfig(gitRepoCache, stackYamlPath, env.InfraRepo)
		if err != nil {
			return nil, err
		}

		envConfigs[env.Name] = stackConfig
	}

	return envConfigs, nil
}
