package server

import (
	"encoding/base32"
	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_MustUser(t *testing.T) {
	store := store.NewTest()

	router := SetupRouter(
		&config.Config{},
		store,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/artifacts")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "should return 401 without an access_token")

	resp, err = http.Get(server.URL + "/api/artifacts?access_token=gibberish")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "should return 401 with a gibberish token")

	user := &model.User{
		Login: "admin",
		Secret: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
		Admin: true,
	}
	err = store.CreateUser(user)
	assert.Nil(t, err)

	tokenInstance := token.New(token.UserToken, user.Login)
	tokenStr, err := tokenInstance.Sign(user.Secret)
	assert.Nil(t, err)

	user = &model.User{
		Login: "user",
		Secret: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
	}
	err = store.CreateUser(user)
	assert.Nil(t, err)

	tokenInstance = token.New(token.UserToken, user.Login)
	tokenStr, err = tokenInstance.Sign(user.Secret)
	assert.Nil(t, err)

	resp, err = http.Get(server.URL + "/api/artifacts?access_token=" + tokenStr)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "should authorize a user with token")
}
