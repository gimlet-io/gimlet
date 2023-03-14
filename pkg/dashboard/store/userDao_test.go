package store

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"

	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	user := model.User{
		Login:            "aLogin",
		AccessToken:      "aGithubToken",
		RefreshToken:     "refreshToken",
		Repos:            []string{"first", "second"},
		FavoriteRepos:    []string{"first", "second"},
		FavoriteServices: []string{"first", "second"},
		Admin:            true,
	}

	err := s.CreateUser(&user)
	assert.Nil(t, err)

	_, err = s.User("noSuchLogin")
	assert.NotNil(t, err)

	u, err := s.User("aLogin")
	assert.Nil(t, err)
	assert.Equal(t, user.Login, u.Login)
	assert.Equal(t, user.AccessToken, u.AccessToken)
	assert.Equal(t, user.RefreshToken, u.RefreshToken)
	assert.Equal(t, user.Repos, u.Repos)
	assert.Equal(t, user.FavoriteRepos, u.FavoriteRepos)
	assert.Equal(t, user.FavoriteServices, u.FavoriteServices)
	assert.Equal(t, user.Admin, u.Admin)

	users, err := s.Users()
	assert.Nil(t, err)
	assert.Equal(t, len(users), 1)

	err = s.DeleteUser("aLogin")
	assert.Nil(t, err)

	users, err = s.Users()
	assert.Nil(t, err)
	assert.Equal(t, len(users), 0)
}

func TestUserWithoutEncryption(t *testing.T) {
	encryptionKey := ""
	encryptionKeyNew := ""
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	user := model.User{
		Login:            "aLogin",
		AccessToken:      "aGithubToken",
		RefreshToken:     "refreshToken",
		Repos:            []string{"first", "second"},
		FavoriteRepos:    []string{"first", "second"},
		FavoriteServices: []string{"first", "second"},
		Admin:            true,
	}

	err := s.CreateUser(&user)
	assert.Nil(t, err)

	_, err = s.User("noSuchLogin")
	assert.NotNil(t, err)

	u, err := s.User("aLogin")
	assert.Nil(t, err)
	assert.Equal(t, user.Login, u.Login)
	assert.Equal(t, user.AccessToken, u.AccessToken)
	assert.Equal(t, user.RefreshToken, u.RefreshToken)
	assert.Equal(t, user.Repos, u.Repos)
	assert.Equal(t, user.FavoriteRepos, u.FavoriteRepos)
	assert.Equal(t, user.FavoriteServices, u.FavoriteServices)
	assert.Equal(t, user.Admin, u.Admin)
}
