package gitops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const Dir_RWX_RX_R = 0754

func Releases(
	repo *git.Repository,
	app, env string,
	repoPerEnv bool,
	since, until *time.Time,
	limit int,
	gitRepo string,
	perf *prometheus.HistogramVec,
) ([]*dx.Release, error) {
	releases := []*dx.Release{}

	var path string
	if env == "" {
		return nil, fmt.Errorf("env is mandatory")
	}

	t0 := time.Now()

	envPath := env
	if repoPerEnv {
		envPath = ""
	}

	if app != "" {
		path = filepath.Join(envPath, app)
	} else {
		path = envPath
	}

	commits, err := repo.Log(
		&git.LogOptions{
			Since: since,
		},
	)
	if err != nil {
		return nil, err
	}
	commits = nativeGit.NewCommitDirIterFromIter(path, commits, repo)

	err = commits.ForEach(func(c *object.Commit) error {
		if limit != -1 && len(releases) >= limit {
			return fmt.Errorf("%s", "LIMIT")
		}

		if RollbackCommit(c) ||
			DeleteCommit(c) {
			return nil
		}

		releaseFile, err := c.File(filepath.Join(envPath, "release.json"))
		if err != nil {
			releaseFile, err = c.File(filepath.Join(path, "release.json"))
			if err != nil {
				logrus.Debugf("no release file for %s: %s", c.Hash.String(), err)
				return nil
			}
		}

		buf := new(bytes.Buffer)
		reader, err := releaseFile.Blob.Reader()
		if err != nil {
			logrus.Warnf("cannot parse release file for %s: %s", c.Hash.String(), err)
			releases = append(releases, releaseFromCommit(c, app, env))
			return nil
		}

		buf.ReadFrom(reader)
		releaseBytes := buf.Bytes()

		var release *dx.Release
		err = json.Unmarshal(releaseBytes, &release)
		if err != nil {
			logrus.Warnf("cannot parse release file for %s: %s", c.Hash.String(), err)
			//releases = append(releases, releaseFromCommit(c, app, env))
			return nil
		}

		if gitRepo != "" { // gitRepo filter
			if release.Version == nil ||
				release.Version.RepositoryName != gitRepo {
				return nil
			}
		}

		release.Created = c.Committer.When.Unix()
		release.GitopsRef = c.Hash.String()

		rolledBack, err := HasBeenReverted(repo, c, env, app, repoPerEnv)
		if err != nil {
			logrus.Warnf("cannot determine if commit was rolled back %s: %s", c.Hash.String(), err)
			releases = append(releases, releaseFromCommit(c, app, env))
		}
		release.RolledBack = rolledBack

		releases = append(releases, release)
		return nil
	})
	if err != nil &&
		err.Error() != "EOF" &&
		err.Error() != "LIMIT" {
		return nil, err
	}

	perf.WithLabelValues("githelper_releases").Observe(float64(time.Since(t0).Seconds()))
	return releases, nil
}

func Status(
	repo *git.Repository,
	app, env string,
	repoPerEnv bool,
	perf *prometheus.HistogramVec,
) (map[string]*dx.Release, error) {
	t0 := time.Now()
	appReleases := map[string]*dx.Release{}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	fs := worktree.Filesystem

	if env == "" {
		return nil, fmt.Errorf("env is mandatory")
	} else {
		if app != "" {
			path := filepath.Join(env, app)
			if repoPerEnv {
				path = app
			}
			release, err := readAppStatus(fs, path)
			if err != nil {
				return nil, fmt.Errorf("cannot read app status %s: %s", path, err)
			}

			appReleases[app] = release
		} else {
			envPath := env
			if repoPerEnv {
				envPath = ""
			}
			paths, err := fs.ReadDir(envPath)
			if err != nil {
				return nil, fmt.Errorf("cannot list files: %s", err)
			}

			for _, fileInfo := range paths {
				if !fileInfo.IsDir() {
					continue
				}
				path := filepath.Join(envPath, fileInfo.Name())

				release, err := readAppStatus(fs, path)
				if err != nil {
					logrus.Debugf("cannot read app status %s: %s", path, err)
				}

				appReleases[fileInfo.Name()] = release
			}
		}
	}

	perf.WithLabelValues("githelper_status").Observe(float64(time.Since(t0).Seconds()))
	return appReleases, nil
}

func Envs(
	repo *git.Repository,
) ([]string, error) {
	var envs []string

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	fs := worktree.Filesystem

	paths, err := fs.ReadDir("/")
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}

	for _, fileInfo := range paths {
		if !fileInfo.IsDir() {
			continue
		}

		dir := fileInfo.Name()
		_, err := readAppStatus(fs, dir)
		if err == nil {
			envs = append(envs, dir)
		}
	}

	return envs, nil
}

func readAppStatus(fs billy.Filesystem, path string) (*dx.Release, error) {
	var release *dx.Release
	f, err := fs.Open(path + "/release.json")
	if err != nil {
		return nil, err
	}

	releaseBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(releaseBytes, &release)
	defer f.Close()
	return release, err
}

func RollbackCommit(c *object.Commit) bool {
	return strings.Contains(c.Message, "This reverts commit")
}

func DeleteCommit(c *object.Commit) bool {
	return strings.Contains(c.Message, "[GimletD delete]")
}

func HasBeenReverted(
	repo *git.Repository,
	commit *object.Commit,
	env string,
	app string,
	repoPerEnv bool,
) (bool, error) {
	var path string
	if app != "" {
		path = filepath.Join(env, app)
		if repoPerEnv {
			path = app
		}
	} else {
		path = env
		if repoPerEnv {
			path = ""
		}
	}

	commits, err := repo.Log(
		&git.LogOptions{
			Since: &commit.Author.When,
		},
	)
	if err != nil {
		return false, errors.WithMessage(err, "could not walk commits")
	}
	commits = nativeGit.NewCommitDirIterFromIter(path, commits, repo)

	hasBeenReverted := false
	err = commits.ForEach(func(c *object.Commit) error {
		if strings.Contains(c.Message, commit.Hash.String()) {
			hasBeenReverted = true
			return fmt.Errorf("EOF")
		}
		return nil
	})
	if err != nil && err.Error() != "EOF" {
		return false, err
	}

	return hasBeenReverted, nil
}

func releaseFromCommit(c *object.Commit, app string, env string) *dx.Release {
	return &dx.Release{
		App:       app,
		Env:       env,
		Created:   c.Committer.When.Unix(),
		GitopsRef: c.Hash.String(),
	}
}