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

package token

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type SecretFunc func(*Token) (string, error)

const (
	SessToken = "sess"
	UserToken = "user"
	CsrfToken = "csrf"
)

type gimletClaims struct {
	Type      string `json:"type,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Subject   string `json:"sub,omitempty"`
}

func (c gimletClaims) Valid() error {
	if c.Type != SessToken {
		return nil
	}

	registeredClaim := jwt.RegisteredClaims{
		ExpiresAt: &jwt.NumericDate{
			Time: time.Unix(c.ExpiresAt, 0),
		},
		IssuedAt: &jwt.NumericDate{
			Time: time.Unix(c.IssuedAt, 0),
		},
		Subject: c.Subject,
	}

	return registeredClaim.Valid()
}

// SignerAlgo is the default algorithm used to sign JWT tokens.
const SignerAlgo = "HS256"

type Token struct {
	Kind    string
	Subject string
}

// Parse parses a JWT token
func Parse(raw string, fn SecretFunc) (*Token, error) {
	token := &Token{}
	parsed, err := jwt.ParseWithClaims(raw, &gimletClaims{}, keyFunc(token, fn))
	if err != nil {
		return nil, err
	} else if !parsed.Valid {
		return nil, jwt.ValidationError{}
	}
	return token, nil
}

// ParseRequest parses a JWT token from the HTTP request
func ParseRequest(r *http.Request, fn SecretFunc) (*Token, error) {
	// first we attempt to get the token from the
	// authorization header.
	var token = r.Header.Get("Authorization")
	if len(token) != 0 {
		cnt, _ := fmt.Sscanf(token, "Bearer %s", &token)
		if cnt != 0 {
			return Parse(token, fn)
		}
	}
	if len(token) != 0 {
		cnt, _ := fmt.Sscanf(token, "BEARER %s", &token)
		if cnt != 0 {
			return Parse(token, fn)
		}
	}

	// then we attempt to get the token from the
	// access_token url query parameter
	token = r.FormValue("access_token")
	if len(token) != 0 {
		return Parse(token, fn)
	}

	// and finally we attempt to get the token from
	// the user session cookie
	cookie, err := r.Cookie("user_sess")
	if err != nil {
		return nil, err
	}
	return Parse(cookie.Value, fn)
}

func CheckCsrf(r *http.Request, fn SecretFunc) error {

	// get and options requests are always
	// enabled, without CSRF checks.
	switch r.Method {
	case "GET", "OPTIONS", "POST":
		return nil
	}

	// parse the raw CSRF token value and validate
	raw := r.Header.Get("X-CSRF-TOKEN")
	_, err := Parse(raw, fn)
	return err
}

func New(kind string, subject string) *Token {
	return &Token{Kind: kind, Subject: subject}
}

// Sign signs the token using the given secret hash
func (t *Token) Sign(secret string) (string, error) {
	return t.SignExpires(secret, 0)
}

// SignExpires signs the token using the given secret hash
// with an expiration date.
func (t *Token) SignExpires(secret string, exp int64) (string, error) {
	var token *jwt.Token

	if exp > 0 {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"type": t.Kind,
			"iat":  time.Now().Unix(),
			"exp":  float64(exp),
			"sub":  t.Subject,
		})
	} else {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"type": t.Kind,
			"iat":  time.Now().Unix(),
			"sub":  t.Subject,
		})
	}

	return token.SignedString([]byte(secret))
}

func keyFunc(token *Token, fn SecretFunc) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		// validate the correct algorithm is being used
		if t.Method.Alg() != SignerAlgo {
			return nil, jwt.ErrSignatureInvalid
		}

		// extract the subject
		claims := t.Claims.(*gimletClaims)
		token.Kind = claims.Type
		token.Subject = claims.Subject

		// invoke the callback function to retrieve
		// the secret key used to verify
		secret, err := fn(token)
		return []byte(secret), err
	}
}
