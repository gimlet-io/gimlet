package nativeGit

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const gitSSHAddressFormat = "git@github.com:%s.git"
const File_RW_RW_R = 0664
const Dir_RWX_RX_R = 0754

func CloneToFs(rootPath string, repoName string, privateKeyPath string) (string, *git.Repository, error) {
	err := os.MkdirAll(rootPath, Dir_RWX_RX_R)
	if err != nil {
		return "", nil, errors.WithMessage(err, "cannot create folder at $REPO_CACHE_PATH")
	}
	path, err := ioutil.TempDir(rootPath, "gitops-")
	if err != nil {
		return "", nil, errors.WithMessage(err, "get temporary directory")
	}
	url := fmt.Sprintf(gitSSHAddressFormat, repoName)
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyPath, "")
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate public key from private: %s", err.Error())
	}

	opts := &git.CloneOptions{
		URL:  url,
		Auth: publicKeys,
	}

	repo, err := git.PlainClone(path, false, opts)
	return path, repo, err
}

func TmpFsCleanup(path string) error {
	return os.RemoveAll(path)
}

func Push(repo *git.Repository, privateKeyPath string) error {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyPath, "")
	if err != nil {
		return fmt.Errorf("cannot generate public key from private: %s", err.Error())
	}

	err = repo.Push(&git.PushOptions{
		Auth: publicKeys,
	})

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

func RemoteFolderOnBranchWithoutCheckout(repo *git.Repository, branch string, path string) (map[string]string, error) {
	files := map[string]string{}

	head := BranchHeadHash(repo, branch)
	headCommit, err := repo.CommitObject(head)
	if err != nil {
		return files, fmt.Errorf("cannot get head commit: %s", err)
	}

	t, err := headCommit.Tree()
	if err != nil {
		return files, fmt.Errorf("cannot get head tree: %s", err)
	}

	subTree, err := t.Tree(path)
	if err != nil {
		return files, fmt.Errorf("cannot get %s tree: %s", path, err)
	}

	for _, entry := range subTree.Entries {
		f, err := subTree.File(entry.Name)
		if err != nil {
			return files, fmt.Errorf("cannot get file: %s", err)
		}
		contents, err := f.Contents()
		if err != nil {
			return files, fmt.Errorf("cannot get file: %s", err)
		}
		files[entry.Name] = contents
	}

	return files, nil
}

func RemoteFoldersOnBranchWithoutCheckout(repo *git.Repository, branch string, path string) ([]string, error) {
	if branch == "" {
		branch = HeadBranch(repo)
	}

	head := BranchHeadHash(repo, branch)
	headCommit, err := repo.CommitObject(head)
	if err != nil {
		return nil, fmt.Errorf("cannot get head commit: %s", err)
	}

	t, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("cannot get head tree: %s", err)
	}

	if path != "" {
		t, err = t.Tree(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get %s tree: %s", path, err)
		}
	}

	folders := []string{}
	for _, entry := range t.Entries {
		if !entry.Mode.IsFile() {
			folders = append(folders, entry.Name)
		}
	}

	return folders, nil
}

func RemoteContentOnBranchWithoutCheckout(repo *git.Repository, branch string, path string) (string, error) {
	head := BranchHeadHash(repo, branch)
	headCommit, err := repo.CommitObject(head)
	if err != nil {
		return "", fmt.Errorf("cannot get head commit: %s", err)
	}

	t, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("cannot get head tree: %s", err)
	}

	f, err := t.File(path)
	if err != nil {
		return "", fmt.Errorf("cannot get head tree: %s", err)
	}

	return f.Contents()
}

func HeadBranch(repo *git.Repository) string {
	branches := BranchList(repo)
	for _, b := range branches {
		if b == "main" {
			return "main"
		}
	}
	return "master"
}

func BranchList(repo *git.Repository) []string {
	branches := []string{}
	refIter, _ := repo.References()
	refIter.ForEach(func(r *plumbing.Reference) error {
		if r.Name().IsRemote() {
			branch := r.Name().Short()
			branches = append(branches, strings.TrimPrefix(branch, "origin/"))
		}
		return nil
	})

	return branches
}

func BranchHeadHash(repo *git.Repository, branch string) plumbing.Hash {
	var head plumbing.Hash
	refIter, _ := repo.References()
	refIter.ForEach(func(r *plumbing.Reference) error {
		if r.Name().IsRemote() {
			remoteBranch := r.Name().Short()
			remoteBranch = strings.TrimPrefix(remoteBranch, "origin/")
			if remoteBranch == branch {
				head = r.Hash()
			}
		}
		return nil
	})

	return head
}

func Branch(repo *git.Repository, ref string) error {
	b := plumbing.ReferenceName(ref)
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})
	if err != nil {
		return err
	}
	return nil
}
