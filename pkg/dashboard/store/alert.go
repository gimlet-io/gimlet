package store

import (
	"database/sql"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) FiringAlerts() ([]*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectFiringAlerts)
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

func (db *Store) CreateAlert(alert *model.Alert) error {
	return meddler.Insert(db, "alerts", alert)
}

func (db *Store) UpdateAlert(alert *model.Alert) error {
	storedAlert, err := db.Alert(alert.Name, alert.Type)

	if err != nil {
		return err
	}

	storedAlert.Status = alert.Status
	// storedAlert.StatusDesc = alert.StatusDesc
	storedAlert.LastStateChange = time.Now().Unix()
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
