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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	pathArtifact   = "%s/api/artifact"
	pathArtifacts  = "%s/api/artifacts"
	pathReleases   = "%s/api/releases"
	pathStatus     = "%s/api/status"
	pathRollback   = "%s/api/rollback"
	pathDelete     = "%s/api/delete"
	pathEvent      = "%s/api/event"
	pathUser       = "%s/api/user"
	pathGitopsRepo = "%s/api/gitopsRepo"
)

type client struct {
	client *http.Client
	addr   string
}

// New returns a client at the specified url.
func New(uri string) Client {
	return &client{http.DefaultClient, strings.TrimSuffix(uri, "/")}
}

// NewClient returns a client at the specified url.
func NewClient(uri string, cli *http.Client) Client {
	return &client{cli, strings.TrimSuffix(uri, "/")}
}

// SetClient sets the http.Client.
func (c *client) SetClient(client *http.Client) {
	c.client = client
}

// SetAddress sets the server address.
func (c *client) SetAddress(addr string) {
	c.addr = addr
}

// ArtifactPost creates a new user account.
func (c *client) ArtifactPost(in *dx.Artifact) (*dx.Artifact, error) {
	out := new(dx.Artifact)
	uri := fmt.Sprintf(pathArtifact, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// ArtifactsGet creates a new user account.
func (c *client) ArtifactsGet(
	repo, branch string,
	event *dx.GitEvent,
	sourceBranch string,
	sha []string,
	limit, offset int,
	since, until *time.Time,
) ([]*dx.Artifact, error) {
	uri := fmt.Sprintf(pathArtifacts, c.addr)

	var params []string

	if limit != 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if offset != 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}
	if since != nil {
		params = append(params, fmt.Sprintf("since=%s", url.QueryEscape(since.Format(time.RFC3339))))
	}
	if until != nil {
		params = append(params, fmt.Sprintf("until=%s", url.QueryEscape(until.Format(time.RFC3339))))
	}
	if repo != "" {
		params = append(params, fmt.Sprintf("repository=%s", repo))
	}
	if branch != "" {
		params = append(params, fmt.Sprintf("branch=%s", branch))
	}
	if event != nil {
		params = append(params, fmt.Sprintf("event=%s", event))
	}
	if sourceBranch != "" {
		params = append(params, fmt.Sprintf("sourceBranch=%s", sourceBranch))
	}
	if len(sha) != 0 {
		for _, s := range sha {
			params = append(params, fmt.Sprintf("sha=%s", s))
		}
	}

	var paramsStr string
	if len(params) > 0 {
		paramsStr = strings.Join(params, "&")
		paramsStr = "?" + paramsStr
	}

	body, err := c.open(uri+paramsStr, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	if bodyString == "[]" { // json deserializer breaks on empty arrays / objects
		return []*dx.Artifact{}, nil
	}

	var out []*dx.Artifact
	err = json.Unmarshal(bodyBytes, &out)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return []*dx.Artifact{}, nil
	}

	return out, err
}

// ReleasesGet creates a new user account.
func (c *client) ReleasesGet(
	app string,
	env string,
	limit, offset int,
	gitRepo string,
	since, until *time.Time,
) ([]*dx.Release, error) {
	uri := fmt.Sprintf(pathReleases, c.addr)

	var params []string

	if limit != 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if offset != 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}
	if since != nil {
		params = append(params, fmt.Sprintf("since=%s", url.QueryEscape(since.Format(time.RFC3339))))
	}
	if until != nil {
		params = append(params, fmt.Sprintf("until=%s", url.QueryEscape(until.Format(time.RFC3339))))
	}
	if app != "" {
		params = append(params, fmt.Sprintf("app=%s", app))
	}
	if env != "" {
		params = append(params, fmt.Sprintf("env=%s", env))
	}
	if gitRepo != "" {
		params = append(params, fmt.Sprintf("git-repo=%s", gitRepo))
	}

	var paramsStr string
	if len(params) > 0 {
		paramsStr = strings.Join(params, "&")
		paramsStr = "?" + paramsStr
	}

	body, err := c.open(uri+paramsStr, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	if bodyString == "[]" { // json deserializer breaks on empty arrays / objects
		return []*dx.Release{}, nil
	}

	var out []*dx.Release
	err = json.Unmarshal(bodyBytes, &out)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return []*dx.Release{}, nil
	}

	return out, err
}

// StatusGet returns release status for all apps in an env
func (c *client) StatusGet(
	app string,
	env string,
) (map[string]*dx.Release, error) {
	uri := fmt.Sprintf(pathStatus, c.addr)

	var params []string
	if app != "" {
		params = append(params, fmt.Sprintf("app=%s", app))
	}
	if env != "" {
		params = append(params, fmt.Sprintf("env=%s", env))
	}

	var paramsStr string
	if len(params) > 0 {
		paramsStr = strings.Join(params, "&")
		paramsStr = "?" + paramsStr
	}

	body, err := c.open(uri+paramsStr, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	if bodyString == "[]" { // json deserializer breaks on empty arrays / objects
		return map[string]*dx.Release{}, nil
	}

	var out map[string]*dx.Release
	err = json.Unmarshal(bodyBytes, &out)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return map[string]*dx.Release{}, nil
	}

	return out, err
}

// ReleasesPost releases the given artifact to the given environment
func (c *client) ReleasesPost(request dx.ReleaseRequest) (string, error) {
	uri := fmt.Sprintf(pathReleases, c.addr)
	result := new(map[string]interface{})
	err := c.post(uri, request, result)
	if err != nil {
		return "", err
	}
	res := *result
	return res["id"].(string), nil
}

// RollbackPost rolls back to a specific gitops commit
func (c *client) RollbackPost(env string, app string, targetSHA string) (string, error) {
	uri := fmt.Sprintf(pathRollback+"?env=%s&app=%s&sha=%s", c.addr, env, app, targetSHA)
	result := new(map[string]interface{})
	err := c.post(uri, nil, result)
	if err != nil {
		return "", err
	}
	res := *result
	return res["id"].(string), nil
}

// DeletePost deletes an application in an env
func (c *client) DeletePost(env string, app string) error {
	uri := fmt.Sprintf(pathDelete+"?env=%s&app=%s", c.addr, env, app)
	result := new(map[string]interface{})
	err := c.post(uri, nil, result)
	if err != nil {
		return err
	}
	return nil
}

// TrackGet gets the status of an event
func (c *client) TrackGet(trackingID string) (*dx.ReleaseStatus, error) {
	uri := fmt.Sprintf(pathEvent, c.addr)

	result := new(dx.ReleaseStatus)
	err := c.get(uri+"?id="+trackingID, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UserGet returns the user with the given login name
func (c *client) UserGet(login string, withToken bool) (*model.User, error) {
	uri := fmt.Sprintf(pathUser, c.addr)

	tokenClause := "?withToken=false"
	if withToken {
		tokenClause = "?withToken=true"
	}

	user := new(model.User)
	err := c.get(uri+"/"+login+tokenClause, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UserPost creates a user
func (c *client) UserPost(toSave *model.User) (*model.User, error) {
	uri := fmt.Sprintf(pathUser, c.addr)
	createdUser := new(model.User)
	err := c.post(uri, toSave, createdUser)
	if err != nil {
		return nil, err
	}
	return createdUser, nil
}

type GitopsRepoResult struct {
	GitopsRepo string `json:"gitopsRepo"`
}

// GitopsRepoGet returns the configured gitops repo name
func (c *client) GitopsRepoGet() (string, error) {
	uri := fmt.Sprintf(pathGitopsRepo, c.addr)

	gitopsRepo := new(GitopsRepoResult)
	err := c.get(uri, gitopsRepo)
	if err != nil {
		return "", err
	}

	return gitopsRepo.GitopsRepo, nil
}

func (c *client) get(rawURL string, out interface{}) error {
	return c.do(rawURL, "GET", nil, out)
}

func (c *client) post(rawURL string, in, out interface{}) error {
	return c.do(rawURL, "POST", in, out)
}

func (c *client) put(rawURL string, in, out interface{}) error {
	return c.do(rawURL, "PUT", in, out)
}

func (c *client) patch(rawURL string, in, out interface{}) error {
	return c.do(rawURL, "PATCH", in, out)
}

func (c *client) delete(rawURL string) error {
	return c.do(rawURL, "DELETE", nil, nil)
}

func (c *client) do(rawURL, method string, in, out interface{}) error {
	body, err := c.open(rawURL, method, in)
	if err != nil {
		return err
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	return json.Unmarshal(bodyBytes, &out)
}

func (c *client) open(rawURL, method string, in interface{}) (io.ReadCloser, error) {
	uri, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	if in != nil {
		decoded, decodeErr := json.Marshal(in)
		if decodeErr != nil {
			return nil, decodeErr
		}
		buf := bytes.NewBuffer(decoded)
		req.Body = ioutil.NopCloser(buf)
		req.ContentLength = int64(len(decoded))
		req.Header.Set("Content-Length", strconv.Itoa(len(decoded)))
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > http.StatusPartialContent {
		defer resp.Body.Close()
		out, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(out))
	}
	return resp.Body, nil
}
