// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package genericScm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gimlet-io/go-scm/scm"
)

// expiryDelta determines how earlier a token should be considered
// expired than its actual expiration time. It is used to avoid late
// expirations due to client-server time mismatches.
const expiryDelta = time.Minute

// Refresher is an http.RoundTripper that refreshes oauth
// tokens, wrapping a base RoundTripper and refreshing the
// token if expired.
//
// IMPORTANT the Refresher is NOT safe for concurrent use
// by multiple goroutines.
type Refresher struct {
	ClientID     string
	ClientSecret string
	Endpoint     string

	Source scm.TokenSource
	Client *http.Client

	tokenUpdater func(token *scm.Token)
}

// Token returns a token. If the token is missing or
// expired, the token is refreshed.
func (t *Refresher) Token(ctx context.Context) (*scm.Token, error) {
	token, err := t.Source.Token(ctx)
	if err != nil {
		return nil, err
	}
	if !expired(token) {
		return token, nil
	}
	err = t.Refresh(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Refresh refreshes the expired token.
func (t *Refresher) Refresh(token *scm.Token) error {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", token.Refresh)

	reader := strings.NewReader(
		values.Encode(),
	)
	req, err := http.NewRequest("POST", t.Endpoint, reader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.ClientID, t.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := t.client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		out := new(tokenError)
		err = json.NewDecoder(res.Body).Decode(out)
		if err != nil {
			return err
		}
		return out
	}

	buf, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return err
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	res.Body = rdr2

	out := new(tokenGrant)
	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		buf := new(strings.Builder)
		io.Copy(buf, rdr1)
		return fmt.Errorf("could not parse token response %s: %s", err, buf.String())
	}

	token.Token = out.Access
	token.Refresh = out.Refresh
	token.Expires = time.Now().Add(
		time.Duration(out.Expires) * time.Second,
	)

	if t.tokenUpdater != nil {
		t.tokenUpdater(token)
	}

	return nil
}

// client returns the http transport. If no base client
// is configured, the default client is returned.
func (t *Refresher) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return http.DefaultClient
}

// expired reports whether the token is expired.
func expired(token *scm.Token) bool {
	if len(token.Refresh) == 0 {
		return false
	}
	if token.Expires.IsZero() && len(token.Token) != 0 {
		return false
	}
	return token.Expires.Add(-expiryDelta).
		Before(time.Now())
}

// tokenGrant is the token returned by the token endpoint.
type tokenGrant struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Expires int64  `json:"expires_in"`
}

// tokenError is the error returned when the token endpoint
// returns a non-2XX HTTP status code.
type tokenError struct {
	Code    string `json:"error"`
	Message string `json:"error_description"`
}

func (t *tokenError) Error() string {
	return t.Message
}
