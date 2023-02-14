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

// TODO update the status for the alert to 'Firing'
func (db *Store) UpdateAlert(alert *model.Alert) error {
	// storedAlert, err := db.Alert(name, alertType)
	// if err != nil {
	// 	return err
	// }

	// storedAlert.Status = "Fired"
	// return meddler.Update(db, "alerts", storedAlert)
	return nil
}

func (db *Store) PendingAlerts() ([]*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPendingAlerts)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, stmt)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}
