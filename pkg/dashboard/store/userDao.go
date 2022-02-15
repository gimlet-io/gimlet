package store

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

// User gets a user by its login name
func (db *Store) User(login string) (*model.User, error) {
	stmt := sql.Stmt(db.driver, sql.SelectUserByLogin)
	data := new(model.User)
	err := meddler.QueryRow(db, data, stmt, login)
	return data, err
}

// CreateUser stores a new user in the database
func (db *Store) CreateUser(user *model.User) error {
	return meddler.Insert(db, "users", user)
}

// UpdateUser updates a user in the database
func (db *Store) UpdateUser(user *model.User) error {
	return meddler.Update(db, "users", user)
}
