package store

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) CreateEnvironment(environment *model.Environment) error {
	return meddler.Insert(db, "environments", environment)
}

func (db *Store) UpdateEnvironment(env *model.Environment) error {
	return meddler.Update(db, "environments", env)
}

func (db *Store) GetEnvironments() ([]*model.Environment, error) {
	stmt := sql.Stmt(db.driver, sql.SelectEnvironments)
	data := []*model.Environment{}
	err := meddler.QueryAll(db, &data, stmt)
	return data, err
}

func (db *Store) GetEnvironment(name string) (*model.Environment, error) {
	stmt := sql.Stmt(db.driver, sql.SelectEnvironment)
	data := new(model.Environment)
	err := meddler.QueryRow(db, data, stmt, name)
	return data, err
}

func (db *Store) DeleteEnvironment(name string) error {
	stmt := sql.Stmt(db.driver, sql.DeleteEnvironment)
	_, err := db.Exec(stmt, name)

	return err
}
