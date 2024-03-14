package store

import (
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

// User gets a user by its login name
func (db *Store) User(login string) (*model.User, error) {
	stmt := sql.Stmt(db.driver, sql.SelectUserByLogin)
	data := new(model.User)
	err := meddler.QueryRow(db, data, stmt, login)
	return data, err
}

// Users returns all users in the database
func (db *Store) Users() ([]*model.User, error) {
	stmt := sql.Stmt(db.driver, sql.SelectAllUser)
	var data []*model.User
	err := meddler.QueryAll(db, &data, stmt)
	return data, err
}

// Users returns all field from all users in the database, included the encrypted data
func (db *Store) EverytingFromUsers() ([]*model.User, error) {
	stmt := sql.Stmt(db.driver, sql.SelectEverytingFromUsers)
	var data []*model.User
	err := meddler.QueryAll(db, &data, stmt)
	return data, err
}

// CreateUser stores a new user in the database
func (db *Store) CreateUser(user *model.User) error {
	return meddler.Insert(db, "users", user)
}

// DeleteUser deletes a user in the database
func (db *Store) DeleteUser(login string) error {
	stmt := sql.Stmt(db.driver, sql.DeleteUser)
	_, err := db.Exec(stmt, login)
	return err
}

// UpdateUser updates a user in the database
func (db *Store) UpdateUser(user *model.User) error {
	return meddler.Update(db, "users", user)
}
