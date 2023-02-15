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

func (db *Store) SaveOrUpdateAlert(alert *model.Alert) error {
	storedAlert, err := db.Alert(alert.Name, alert.Type)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return meddler.Insert(db, "alerts", alert)
		default:
			return err
		}
	}

	storedAlert.DeploymentName = alert.DeploymentName
	storedAlert.Status = alert.Status
	storedAlert.StatusDesc = alert.StatusDesc
	storedAlert.LastStateChange = alert.LastStateChange
	storedAlert.Count = alert.Count
	return meddler.Update(db, "alerts", storedAlert)
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
