package server

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
)

func saveUser(w http.ResponseWriter, r *http.Request) {
	var usernameToSave string
	err := json.NewDecoder(r.Body).Decode(&usernameToSave)
	if err != nil {
		logrus.Errorf("cannot decode user name to save: %s", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	ctx := r.Context()
	store := ctx.Value("store").(*store.Store)

	user := &model.User{
		Login:  usernameToSave,
		Secret: base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
	}

	err = store.CreateUser(user)
	if err != nil {
		logrus.Errorf("cannot creat user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	token := token.New(token.UserToken, user.Login)
	tokenStr, err := token.Sign(user.Secret)
	if err != nil {
		logrus.Errorf("couldn't create user token %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// token is not saved as it is JWT
	user.Token = tokenStr

	userString, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("cannot serialize user %s: %s", user.Login, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(userString)
}

type App struct {
	Name     string        `json:"name"`
	Releases []*dx.Release `json:"releases"`
}

type Env struct {
	Name string `json:"name"`
	Apps []*App `json:"apps"`
}

func insertIntoRolloutHistory(rolloutHistory []*Env, release *dx.Release, perAppLimit int) []*Env {
	for _, env := range rolloutHistory {
		if env.Name == release.Env {
			for _, app := range env.Apps {
				if app.Name == release.App {
					if len(app.Releases) < perAppLimit {
						app.Releases = append(app.Releases, release)
					}
					return rolloutHistory
				}
			}

			env.Apps = append(env.Apps, &App{
				Name:     release.App,
				Releases: []*dx.Release{release},
			})
			return rolloutHistory
		}
	}

	rolloutHistory = append(rolloutHistory, &Env{
		Name: release.Env,
		Apps: []*App{
			{
				Name:     release.App,
				Releases: []*dx.Release{release},
			},
		},
	})

	return rolloutHistory
}

type ByCreated []*dx.Release

func (a ByCreated) Len() int           { return len(a) }
func (a ByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreated) Less(i, j int) bool { return a[i].Created < a[j].Created }

func orderRolloutHistoryFromAscending(rolloutHistory []*Env) []*Env {
	orderedRolloutHistory := []*Env{}

	for _, env := range rolloutHistory {
		orderedApps := []*App{}

		for _, app := range env.Apps {
			sort.Sort(ByCreated(app.Releases))
			orderedApps = append(orderedApps, app)
		}

		env.Apps = orderedApps
		orderedRolloutHistory = append(orderedRolloutHistory, env)
	}

	return orderedRolloutHistory
}

func decorateCommitsWithGimletArtifacts(commits []*Commit, store *store.Store, repo *git.Repository, owner string, repoName string) ([]*Commit, error) {
	var hashes []string
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	err := generateFakeArtifactsForCommits(hashes, store, repo, repoName)
	if err != nil {
		return nil, err
	}

	return doDecorateCommitsWithGimletArtifacts(hashes, commits, store)
}

func doDecorateCommitsWithGimletArtifacts(hashes []string, commits []*Commit, store *store.Store) ([]*Commit, error) {
	events, err := store.Artifacts(
		"", "",
		nil,
		"",
		hashes,
		0, 0, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get artifacts: %s", err)
	}

	artifacts := []*dx.Artifact{}
	for _, a := range events {
		artifact, err := model.ToArtifact(a)
		if err != nil {
			return nil, fmt.Errorf("cannot deserialize artifact: %s", err)
		}
		artifacts = append(artifacts, artifact)
	}

	artifactsBySha := map[string]*dx.Artifact{}
	for _, a := range artifacts {
		artifactsBySha[a.Version.SHA] = a
	}

	var decoratedCommits []*Commit
	for _, c := range commits {
		if artifact, ok := artifactsBySha[c.SHA]; ok {
			for _, targetEnv := range artifact.Environments {
				targetEnv.ResolveVars(artifact.CollectVariables())
				if c.DeployTargets == nil {
					c.DeployTargets = []*api.DeployTarget{}
				}
				c.DeployTargets = append(c.DeployTargets, &api.DeployTarget{
					App:        targetEnv.App,
					Env:        targetEnv.Env,
					Tenant:     targetEnv.Tenant.Name,
					ArtifactId: artifact.ID,
				})
			}
		}
		decoratedCommits = append(decoratedCommits, c)
	}

	return decoratedCommits, nil
}

func generateFakeArtifactsForCommits(hashes []string, store *store.Store, repo *git.Repository, repoName string) error {
	for _, hash := range hashes {
		key := fmt.Sprintf("%s-%s", model.CommitArtifactsGenerated, hash)
		_, err := store.KeyValue(key)
		if err == nil {
			continue
		}

		err = generateFakeArtifact(hash, store, repo, repoName)
		if err != nil {
			return err
		}

		store.SaveKeyValue(&model.KeyValue{
			Key: key,
		})
	}

	return nil
}

func generateFakeArtifact(hash string, store *store.Store, repo *git.Repository, repoName string) error {
	manifests, err := gitops.Manifests(repo, hash)
	if err != nil {
		return err
	}

	manifestsThatNeedFakeArtifact := []*dx.Manifest{}
	for _, m := range manifests {
		strategy := gitops.ExtractImageStrategy(m)

		if strategy == "static" ||
			strategy == "static-site" ||
			strategy == "buildpacks" {
			manifestsThatNeedFakeArtifact = append(manifestsThatNeedFakeArtifact, m)
		}
	}

	err = doGenerateFakeArtifact(hash, manifestsThatNeedFakeArtifact, store, repoName, repo)
	if err != nil {
		return err
	}

	return nil
}

func doGenerateFakeArtifact(
	hash string,
	manifests []*dx.Manifest,
	store *store.Store,
	repoName string,
	repo *git.Repository,
) error {
	owner, name := scm.Split(repoName)
	version, err := gitops.Version(owner, name, repo, hash)
	if err != nil {
		return err
	}

	artifact := &dx.Artifact{
		ID:           fmt.Sprintf("%s-%s", repoName, uuid.New().String()),
		Created:      time.Now().Unix(),
		Fake:         true,
		Environments: manifests,
		Version:      *version,
		Vars: map[string]string{
			"SHA": hash,
		},
	}

	event, err := model.ToEvent(*artifact)
	if err != nil {
		return err
	}

	_, err = store.CreateEvent(event)
	if err != nil {
		return err
	}

	return nil
}
