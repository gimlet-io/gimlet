package worker

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ReleaseStateWorker struct {
	RepoCache     *nativeGit.RepoCache
	Releases      *prometheus.GaugeVec
	Perf          *prometheus.HistogramVec
	Store         *store.Store
	DynamicConfig *dynamicconfig.DynamicConfig
}

func (w *ReleaseStateWorker) Run() {
	for {
		t0 := time.Now()

		envsInDB, err := w.Store.GetEnvironments()
		if err != nil {
			logrus.Warnf("could not get envs: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}

		for _, env := range envsInDB {
			err = processRepo(
				env,
				w.Releases,
				w.Perf,
				w.RepoCache,
				w.DynamicConfig.ScmURL(),
			)
			if err != nil {
				logrus.Warnf("could not process state of %s gitops repo", env.Name)
			}
		}

		w.Perf.WithLabelValues("releaseState_run").Observe(time.Since(t0).Seconds())
		time.Sleep(30 * time.Second)
	}
}

func processRepo(
	processEnv *model.Environment,
	releases *prometheus.GaugeVec,
	perf *prometheus.HistogramVec,
	repoCache *nativeGit.RepoCache,
	scmUrl string,
) error {
	t0 := time.Now()
	perf.WithLabelValues("releaseState_clone").Observe(time.Since(t0).Seconds())

	var envs []string
	var err error
	if processEnv.RepoPerEnv {
		envs = []string{processEnv.Name}
	} else {
		repoCache.PerformAction(processEnv.AppsRepo, func(repo *git.Repository) {
			envs, err = gitops.Envs(repo)
		})
		if err != nil {
			return fmt.Errorf("cannot get envs: %s", err)
		}
	}

	releases.Reset()
	for _, env := range envs {
		t1 := time.Now()
		var appReleases map[string]*dx.Release
		var err error
		repoCache.PerformAction(processEnv.AppsRepo, func(repo *git.Repository) {
			appReleases, err = gitops.Status(repo, "", env, processEnv.RepoPerEnv, perf)
		})
		if err != nil {
			logrus.Errorf("cannot get status: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		perf.WithLabelValues("releaseState_appReleases").Observe(time.Since(t1).Seconds())

		envPath := env
		if processEnv.RepoPerEnv {
			envPath = ""
		}

		for app, release := range appReleases {
			t2 := time.Now()
			var commit *object.Commit
			var err error
			repoCache.PerformAction(processEnv.AppsRepo, func(repo *git.Repository) {
				commit, err = lastCommitThatTouchedAFile(repo, filepath.Join(envPath, app))
			})
			if err != nil {
				logrus.Errorf("cannot find last commit: %s", err)
				time.Sleep(30 * time.Second)
				continue
			}
			if commit == nil {
				continue
			}
			perf.WithLabelValues("releaseState_appRelease").Observe(time.Since(t2).Seconds())

			gitopsRef := fmt.Sprintf(scmUrl+"/%s/commit/%s", processEnv.AppsRepo, commit.Hash.String())
			created := commit.Committer.When

			if release != nil {
				releases.WithLabelValues(
					env,
					app,
					release.Version.URL,
					release.Version.Message,
					gitopsRef,
					created.Format(time.RFC3339),
				).Set(1.0)
			} else {
				releases.WithLabelValues(
					env,
					app,
					"",
					"",
					gitopsRef,
					created.Format(time.RFC3339),
				).Set(1.0)
			}
		}
	}

	return nil
}

func lastCommitThatTouchedAFile(repo *git.Repository, path string) (*object.Commit, error) {
	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}
	commits = nativeGit.NewCommitDirIterFromIter(path, commits, repo)

	var commit *object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		commit = c
		return fmt.Errorf("%s", "FOUND")
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "FOUND" {
		return nil, err
	}

	return commit, nil
}
