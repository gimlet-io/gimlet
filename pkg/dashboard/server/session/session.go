// Original work Copyright 2018 Drone.IO Inc.
// Modified work Copyright 2019 Laszlo Fogas
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

package session

import (
	"context"
	"net/http"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/server/token"
	"github.com/sirupsen/logrus"
)

func SetUser() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var user *model.User

			ctx := r.Context()
			store := ctx.Value("store").(*store.Store)

			t, err := token.ParseRequest(r, func(t *token.Token) (string, error) {
				var err error
				user, err = store.User(t.Subject)
				if err != nil {
					return "", err
				}
				return user.Secret, err
			})
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))

				// if this is a session token (ie not the API token)
				// this means the user is accessing with a web browser,
				// so we should implement CSRF protection measures.
				if t.Kind == token.SessToken {
					err = token.CheckCsrf(r, func(t *token.Token) (string, error) {
						return user.Secret, nil
					})
					// if csrf token validation fails, exit immediately
					// with a not authorized error.
					if err != nil {
						logrus.Warnf("csrf token error: %s", err)
						http.Error(w, http.StatusText(401), 401)
						return
					}
				}
			} else {
				logrus.Warnf("could not set user: %s", err)
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// SetCSRF sets the X-CSRF-TOKEN header with a signed token to prevent CSRF
func SetCSRF() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, userSet := ctx.Value("user").(*model.User)
			if userSet {
				var csrf string
				if user != nil {
					csrf, _ = token.New(
						token.CsrfToken,
						user.Login,
						"",
					).Sign(user.Secret)
				}
				w.Header().Set("X-CSRF-TOKEN", csrf)
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// MustUser makes sure there is an authenticated user set
func MustUser() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			_, userSet := ctx.Value("user").(*model.User)
			if !userSet {
				if r.URL.Path == "/settings/installed" {
					http.Redirect(w, r, "/auth?"+r.URL.RawQuery, http.StatusSeeOther)
				}
				http.Error(w, http.StatusText(401), 401)
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}

// MustAdmin makes sure there is an authenticated user set and she is admin
func MustAdmin() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, userSet := ctx.Value("user").(*model.User)
			if !userSet {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			} else if user.Admin {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, http.StatusText(http.StatusForbidden)+" admin user is required", http.StatusForbidden)
			}
		}
		return http.HandlerFunc(fn)
	}
}
