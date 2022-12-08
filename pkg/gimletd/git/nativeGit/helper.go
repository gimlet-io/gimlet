package nativeGit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const File_RW_RW_R = 0664
const Dir_RWX_RX_R = 0754

func Push(repo *git.Repository, privateKeyPath string) error {
	t0 := time.Now().UnixNano()
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyPath, "")
	if err != nil {
		return fmt.Errorf("cannot generate public key from private: %s", err.Error())
	}
	logrus.Infof("Reading public key took %d", (time.Now().UnixNano()-t0)/1000/1000)

	t0 = time.Now().UnixNano()
	err = repo.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	logrus.Infof("Actual push took %d", (time.Now().UnixNano()-t0)/1000/1000)

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func PushWithToken(repo *git.Repository, accessToken string) error {
	err := repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: accessToken,
		},
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func NothingToCommit(repo *git.Repository) (bool, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, err
	}

	return status.IsClean(), nil
}

func Commit(repo *git.Repository, message string) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	sha, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Gimlet CLI",
			Email: "cli@gimlet.io",
			When:  time.Now(),
		},
	})

	if err != nil {
		return "", err
	}

	return sha.String(), nil
}

func NativeRevert(repoPath string, sha string) error {
	return execCommand(repoPath, "git", "revert", sha)
}

func NativePush(repoPath string, privateKeyPath string, branch string) error {
	sshCommand := fmt.Sprintf("ssh -i %s", privateKeyPath)
	err := execCommand(repoPath, "git", "config", "core.sshCommand", sshCommand)
	if err != nil {
		return err
	}
	err = execCommand(repoPath, "git", "pull", "--rebase")
	if err != nil {
		return err
	}
	return execCommand(repoPath, "git", "push", "origin", branch)
}

func execCommand(rootPath string, cmdName string, args ...string) error {
	cmd := exec.CommandContext(context.TODO(), cmdName, args...)
	cmd.Dir = rootPath
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithMessage(err, "get stdout pipe for command")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithMessage(err, "get stderr pipe for command")
	}
	err = cmd.Start()
	if err != nil {
		return errors.WithMessage(err, "start command")
	}

	stdoutData, err := ioutil.ReadAll(stdout)
	if err != nil {
		return errors.WithMessage(err, "read stdout data of command")
	}
	stderrData, err := ioutil.ReadAll(stderr)
	if err != nil {
		return errors.WithMessage(err, "read stderr data of command")
	}

	err = cmd.Wait()
	logrus.Infof("git/commit: exec command '%s %s': stdout: %s", cmdName, strings.Join(args, " "), stdoutData)
	logrus.Infof("git/commit: exec command '%s %s': stderr: %s", cmdName, strings.Join(args, " "), stderrData)
	if err != nil {
		return fmt.Errorf("cannot execute command %s: %s", err.Error(), stderrData)
	}

	return nil
}

func DelDir(repo *git.Repository, path string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Filesystem.Stat(path)
	if err != nil {
		return nil
	}

	files, err := worktree.Filesystem.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			DelDir(repo, file.Name())
		}

		_, err = worktree.Remove(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
	}

	_, err = worktree.Remove(path)

	return err
}

func StageFolder(repo *git.Repository, folder string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	return worktree.AddWithOptions(&git.AddOptions{
		Glob: folder + "/*",
	})
}

func CommitFilesToGit(
	repo *git.Repository,
	files map[string]string,
	env string,
	app string,
	repoPerEnv bool,
	message string,
	releaseString string,
) (string, error) {
	empty, err := NothingToCommit(repo)
	if err != nil {
		return "", fmt.Errorf("cannot get git state %s", err)
	}
	if !empty {
		return "", fmt.Errorf("there are staged changes in the gitops repo. Commit them first then try again")
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("cannot get worktree %s", err)
	}

	rootPath := filepath.Join(env, app)
	if repoPerEnv {
		rootPath = app
	}

	// first delete, then recreate app dir
	// to remove stale template files
	err = DelDir(repo, rootPath)
	if err != nil {
		return "", fmt.Errorf("cannot del dir: %s", err)
	}
	err = w.Filesystem.MkdirAll(rootPath, Dir_RWX_RX_R)
	if err != nil {
		return "", fmt.Errorf("cannot create dir %s", err)
	}

	for path, content := range files {
		if !strings.HasSuffix(content, "\n") {
			content = content + "\n"
		}

		err = stageFile(w, content, filepath.Join(rootPath, filepath.Base(path)))
		if err != nil {
			return "", fmt.Errorf("cannot stage file %s", err)
		}
	}

	if releaseString != "" {
		if !strings.HasSuffix(releaseString, "\n") {
			releaseString = releaseString + "\n"
		}

		envReleaseJsonPath := env
		if repoPerEnv {
			envReleaseJsonPath = ""
		}

		err = stageFile(w, releaseString, filepath.Join(envReleaseJsonPath, "release.json"))
		if err != nil {
			return "", fmt.Errorf("cannot stage file %s", err)
		}
		err = stageFile(w, releaseString, filepath.Join(rootPath, "release.json"))
		if err != nil {
			return "", fmt.Errorf("cannot stage file %s", err)
		}
	}

	empty, err = NothingToCommit(repo)
	if err != nil {
		return "", err
	}
	if empty {
		return "", nil
	}

	gitMessage := fmt.Sprintf("[Gimlet] %s/%s %s", env, app, message)
	return Commit(repo, gitMessage)
}

func stageFile(worktree *git.Worktree, content string, path string) error {
	createdFile, err := worktree.Filesystem.Create(path)
	if err != nil {
		return err
	}
	_, err = createdFile.Write([]byte(content))
	if err != nil {
		return err
	}
	err = createdFile.Close()
	if err != nil {
		return err
	}

	_, err = worktree.Add(path)
	return err
}

func Branch(repo *git.Repository, ref string) error {
	b := plumbing.ReferenceName(ref)
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{Create: false, Force: false, Branch: b})
	if err != nil {
		return err
	}
	return nil
}

// Content returns the content of a file
func Content(repo *git.Repository, path string) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	f, err := worktree.Filesystem.Open(path)
	if err != nil {
		return "", nil
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Folder returns the file contents of a folder (non-recursive)
func Folder(repo *git.Repository, path string) (map[string]string, error) {
	files := map[string]string{}

	worktree, err := repo.Worktree()
	if err != nil {
		return files, err
	}

	fileInfos, err := worktree.Filesystem.ReadDir(path)
	if err != nil {
		return files, err
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		f, err := worktree.Filesystem.Open(filepath.Join(path, fileInfo.Name()))
		if err != nil {
			return files, nil
		}
		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return files, err
		}

		files[fileInfo.Name()] = string(content)
	}

	return files, nil
}

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
	commits = NewCommitDirIterFromIter(path, commits, repo)

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
	commits = NewCommitDirIterFromIter(path, commits, repo)

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
