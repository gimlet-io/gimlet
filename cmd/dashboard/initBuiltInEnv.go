package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gimlet-io/gimlet-cli/pkg/gitops"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/gorilla/securecookie"
	"github.com/sosedoff/gitkit"
)

func bootstrapBuiltInEnv(
	store *store.Store,
	repoCache *nativeGit.RepoCache,
	gitUser *model.User,
	config *config.Config,
	jwtSecret string,
) error {
	envsInDB, err := store.GetEnvironments()
	if err != nil {
		panic(err)
	}
	for _, env := range envsInDB {
		if env.BuiltIn {
			return nil
		}
	}

	randomFirstName := firstNames[rand.Intn(len(firstNames))]
	randomSecondName := secondNames[rand.Intn(len(secondNames))]
	builtInEnv := &model.Environment{
		Name:       fmt.Sprintf("%s-%s", randomFirstName, randomSecondName),
		InfraRepo:  "builtin/infra",
		AppsRepo:   "builtin/apps",
		BuiltIn:    true,
		RepoPerEnv: true,
	}
	err = store.CreateEnvironment(builtInEnv)
	if err != nil {
		return err
	}

	repo, tmpPath, err := initRepo(fmt.Sprintf("http://%s/%s", config.GitHost, builtInEnv.InfraRepo))
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot get repo: %s", err)
	}

	headBranch, err := nativeGit.HeadBranch(repo)
	if err != nil {
		return fmt.Errorf("cannot get head branch: %s", err)
	}

	opts := gitops.DefaultManifestOpts()
	opts.ShouldGenerateDeployKey = false
	opts.ShouldGenerateBasicAuthSecret = true
	opts.BasicAuthUser = gitUser.Login
	opts.BasicAuthPassword = gitUser.Secret
	opts.GitopsRepoUrl = fmt.Sprintf("http://%s/%s", config.GitHost, builtInEnv.InfraRepo)
	opts.GitopsRepoPath = tmpPath
	opts.Branch = headBranch
	_, _, _, err = gitops.GenerateManifests(opts)
	if err != nil {
		return fmt.Errorf("cannot generate manifest: %s", err)
	}

	err = server.PrepAgentManifests(builtInEnv, tmpPath, repo, config.Host, jwtSecret)
	if err != nil {
		return fmt.Errorf("cannot configure agent: %s", err)
	}

	err = stageCommitAndPush(repo, tmpPath, gitUser.Login, gitUser.Secret, "[Gimlet] Bootstrapping")
	if err != nil {
		return fmt.Errorf("cannot stage commit and push: %s", err)
	}

	repo, tmpPath, err = initRepo(fmt.Sprintf("http://%s/%s", config.GitHost, builtInEnv.AppsRepo))
	defer os.RemoveAll(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot get repo: %s", err)
	}

	opts = gitops.DefaultManifestOpts()
	opts.ShouldGenerateController = false
	opts.ShouldGenerateDependencies = false
	opts.ShouldGenerateDeployKey = false
	opts.ShouldGenerateBasicAuthSecret = true
	opts.BasicAuthUser = gitUser.Login
	opts.BasicAuthPassword = gitUser.Secret
	opts.GitopsRepoUrl = fmt.Sprintf("http://%s/%s", config.GitHost, builtInEnv.AppsRepo)
	opts.GitopsRepoPath = tmpPath
	opts.Branch = headBranch
	_, _, _, err = gitops.GenerateManifests(opts)
	if err != nil {
		return fmt.Errorf("cannot generate manifest: %s", err)
	}

	gimletToken, err := server.PrepNotificationsApiKey(builtInEnv, store)
	if err != nil {
		return fmt.Errorf("couldn't create user token %s", err)
	}

	_, err = gitops.GenerateManifestProviderAndAlert(builtInEnv, tmpPath, config.Host, gimletToken)
	if err != nil {
		return fmt.Errorf("cannot generate notifications manifest: %s", err)
	}

	err = stageCommitAndPush(repo, tmpPath, gitUser.Login, gitUser.Secret, "[Gimlet] Bootstrapping")
	if err != nil {
		return fmt.Errorf("cannot stage commit and push: %s", err)
	}

	return nil
}

func initRepo(url string) (*git.Repository, string, error) {
	tmpPath, _ := ioutil.TempDir("", "gitops-")
	repo, err := git.PlainInit(tmpPath, false)
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	w, err := repo.Worktree()
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	err = nativeGit.StageFile(w, "", "README.md")
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	_, err = nativeGit.Commit(repo, "Init")
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}
	_, err = repo.CreateRemote(&gitConfig.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return nil, tmpPath, fmt.Errorf("cannot init empty repo: %s", err)
	}

	return repo, tmpPath, nil
}

func stageCommitAndPush(repo *git.Repository, tmpPath string, user string, password string, msg string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	// Temporarily staging deleted files to git with a simple CLI command until the
	// following issue is not solved:
	// https://github.com/go-git/go-git/issues/223
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpPath
	err = cmd.Run()
	if err != nil {
		return err
	}

	_, err = nativeGit.Commit(repo, msg)
	if err != nil {
		return err
	}

	err = nativeGit.PushWithBasicAuth(repo, user, password)
	if err != nil {
		return err
	}

	return nil
}

func builtInGitServer(gitUser *model.User, gitRoot string) (http.Handler, error) {
	hooks := &gitkit.HookScripts{}

	service := gitkit.New(gitkit.Config{
		Dir:        gitRoot,
		AutoCreate: true,
		AutoHooks:  true,
		Hooks:      hooks,
		Auth:       true,
	})

	// Here's the user-defined authentication function.
	// If return value is false or error is set, user's request will be rejected.
	// You can hook up your database/redis/cache for authentication purposes.
	service.AuthFunc = func(cred gitkit.Credential, req *gitkit.Request) (bool, error) {
		log.Println("user auth request for repo:", cred.Username, cred.Password, req.RepoName)
		return cred.Username == "git" && cred.Password == gitUser.Secret, nil
	}

	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := service.Setup(); err != nil {
		return nil, err
	}

	return service, nil
}

func setupGitUser(config *config.Config, store *store.Store) (*model.User, error) {
	gitUser, err := store.User("git")

	if err == sql.ErrNoRows {
		gitUser = &model.User{
			Login: "git",
			Secret: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
			Admin: false,
		}
		err = store.CreateUser(gitUser)
		if err != nil {
			return nil, fmt.Errorf("couldn't create user git user %s", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("couldn't list users to create admin user %s", err)
	}

	return gitUser, nil
}
