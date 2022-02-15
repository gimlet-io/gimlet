package client

import (
	"encoding/base32"
	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server"
	"github.com/gimlet-io/gimlet-cli/pkg/server/token"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

import (
	"golang.org/x/oauth2"
)

func Test_artifact(t *testing.T) {
	store := store.NewTest()

	router := server.SetupRouter(&config.Config{}, store, nil, nil, nil)
	server := httptest.NewServer(router)
	defer server.Close()

	user := &model.User{
		Login: "admin",
		Secret: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
	}
	err := store.CreateUser(user)
	assert.Nil(t, err)

	tokenInstance := token.New(token.UserToken, user.Login)
	tokenStr, err := tokenInstance.Sign(user.Secret)
	assert.Nil(t, err)

	config := new(oauth2.Config)
	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: tokenStr,
		},
	)

	client := NewClient(server.URL, auther)

	savedArtifact, err := client.ArtifactPost(&dx.Artifact{
		Version: dx.Version{
			SHA:            "sha",
			RepositoryName: "my-app",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "sha", savedArtifact.Version.SHA)

	artifacts, err := client.ArtifactsGet(
		"", "",
		nil,
		"",
		[]string{},
		0, 0,
		nil, nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(artifacts))
}
