// Copyright 2021 Laszlo Fogas
// Original structure Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"net/http"
	"time"
)

// Client is used to communicate with a Drone server.
type Client interface {
	// SetClient sets the http.Client.
	SetClient(*http.Client)

	// SetAddress sets the server address.
	SetAddress(string)

	// ArtifactPost creates a new artifact.
	ArtifactPost(artifact *dx.Artifact) (*dx.Artifact, error)

	// ArtifactsGet returns all artifacts in the database within the given constraints
	ArtifactsGet(
		repo, branch string,
		event *dx.GitEvent,
		sourceBranch string,
		sha []string,
		limit, offset int,
		since, until *time.Time,
	) ([]*dx.Artifact, error)

	// ReleasesGet returns all releases from the gitops repo within the given constraints
	ReleasesGet(
		app string,
		env string,
		limit, offset int,
		gitRepo string,
		since, until *time.Time,
	) ([]*dx.Release, error)

	// StatusGet returns release status for all apps in an env
	StatusGet(
		app string,
		env string,
	) (map[string]*dx.Release, error)

	// ReleasesPost releases the given artifact to the given environment
	ReleasesPost(request dx.ReleaseRequest) (string, error)

	// RollbackPost rolls back to the given sha
	RollbackPost(env string, app string, targetSHA string) (string, error)

	// DeletePost deletes an application in an env
	DeletePost(env string, app string) error

	// TrackGet returns the state of an event
	TrackGet(trackingID string) (*dx.ReleaseStatus, error)

	// UserGet returns the user with the given login
	UserGet(login string, withToken bool) (*model.User, error)

	// UsersGet returns all users
	UsersGet() ([]*model.User, error)

	// UserPost creates a user
	UserPost(user *model.User) (*model.User, error)

	// GitopsRepoGet returns the configured gitops repo name
	GitopsRepoGet() (string, error)
}
