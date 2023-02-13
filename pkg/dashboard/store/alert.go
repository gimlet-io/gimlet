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

func (db *Store) Alert(name string, alertType string) (*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectAlertByNameAndType)
	alert := new(model.Alert)
	err := meddler.QueryRow(db, alert, stmt, name, alertType)

	return alert, err
}

func (db *Store) SaveAlert(alert *model.Alert) error {
	return meddler.Insert(db, "alerts", alert)
}

func (db *Store) PendingAlertsByType(alertType string) ([]*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPendingAlertsByType)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, stmt, alertType)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}
