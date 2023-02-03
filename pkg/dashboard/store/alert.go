package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Alerts() ([]*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectAllAlerts)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, stmt)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) SaveAlert(alert *model.Alert) error {
	return meddler.Insert(db, "alerts", alert)
}

func (db *Store) DeleteAlert(name string, alertType string) error {
	stmt := queries.Stmt(db.driver, queries.DeleteAlertByNameAndType)
	_, err := db.Exec(stmt, name, alertType)

	return err
}
