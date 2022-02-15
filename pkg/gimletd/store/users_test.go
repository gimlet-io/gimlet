package store

import (
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	user := model.User{
		Login: "aLogin",
	}

	err := s.CreateUser(&user)
	assert.Nil(t, err)

	_, err = s.User("noSuchLogin")
	assert.NotNil(t, err)

	u, err := s.User("aLogin")
	assert.Nil(t, err)
	assert.Equal(t, user.Login, u.Login)

	users, err := s.Users()
	assert.Nil(t, err)
	assert.Equal(t, len(users), 1)

	err = s.DeleteUser("aLogin")
	assert.Nil(t, err)

	users, err = s.Users()
	assert.Nil(t, err)
	assert.Equal(t, len(users), 0)
}
