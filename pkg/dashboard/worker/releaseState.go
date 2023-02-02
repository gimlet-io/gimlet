package worker

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ReleaseStateWorker struct {
	GitopsRepo      string
	RepoCache       *gitops.GitopsRepoCache
	Releases        *prometheus.GaugeVec
	Perf            *prometheus.HistogramVec
	GitopsRepos     map[string]*config.GitopsRepoConfig
	DefaultRepoName string
}

func (w *ReleaseStateWorker) Run() {
	for {
		t0 := time.Now()
		for _, repoConfig := range w.GitopsRepos {
			err := processRepo(
				repoConfig,
				w.Releases,
				w.Perf,
				w.RepoCache,
			)
			if err != nil {
				logrus.Warnf("could not process state of %s gitops repo: %s", repoConfig.GitopsRepo, err)
				time.Sleep(30 * time.Second)
				continue

			}
		}
		if w.DefaultRepoName != "" {
			err := processRepo(
				&config.GitopsRepoConfig{
					GitopsRepo: w.DefaultRepoName,
					RepoPerEnv: false,
				},
				w.Releases,
				w.Perf,
				w.RepoCache,
			)
			if err != nil {
				logrus.Warnf("could not process state of %s gitops repo", w.DefaultRepoName)
			}
		}
		w.Perf.WithLabelValues("releaseState_run").Observe(time.Since(t0).Seconds())
		time.Sleep(30 * time.Second)
	}
}

func processRepo(
	gitopsRepoConfig *config.GitopsRepoConfig,
	releases *prometheus.GaugeVec,
	perf *prometheus.HistogramVec,
	repoCache *gitops.GitopsRepoCache,
) error {
	t0 := time.Now()
	repo := repoCache.InstanceForRead(gitopsRepoConfig.GitopsRepo)
	perf.WithLabelValues("releaseState_clone").Observe(time.Since(t0).Seconds())

	var envs []string
	var err error
	if gitopsRepoConfig.RepoPerEnv {
		envs = []string{gitopsRepoConfig.Env}
	} else {
		envs, err = gitops.Envs(repo)
		if err != nil {
			return fmt.Errorf("cannot get envs: %s", err)
		}
	}

	releases.Reset()
	for _, env := range envs {
		t1 := time.Now()
		appReleases, err := gitops.Status(repo, "", env, gitopsRepoConfig.RepoPerEnv, perf)
		if err != nil {
			logrus.Errorf("cannot get status: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		perf.WithLabelValues("releaseState_appReleases").Observe(time.Since(t1).Seconds())

		envPath := env
		if gitopsRepoConfig.RepoPerEnv {
			envPath = ""
		}

		for app, release := range appReleases {
			t2 := time.Now()
			commit, err := lastCommitThatTouchedAFile(repo, filepath.Join(envPath, app))
			if err != nil {
				logrus.Errorf("cannot find last commit: %s", err)
				time.Sleep(30 * time.Second)
				continue
			}
			if commit == nil {
				continue
			}
			perf.WithLabelValues("releaseState_appRelease").Observe(time.Since(t2).Seconds())

			gitopsRef := fmt.Sprintf("https://github.com/%s/commit/%s", gitopsRepoConfig.GitopsRepo, commit.Hash.String())
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
