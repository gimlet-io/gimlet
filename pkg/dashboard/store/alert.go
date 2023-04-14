package store

import (
	"database/sql"

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

func (db *Store) Alerts(objectName string, objectType string) ([]*model.Alert, error) {
	stmt := queries.Stmt(db.driver, queries.SelectAlertsByNameAndType)

	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, stmt, objectName, objectType)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) CreateAlert(alert *model.Alert) error {
	return meddler.Insert(db, "alerts", alert)
}

func (db *Store) DeleteAlert(id int64) error {
	stmt := queries.Stmt(db.driver, queries.DeleteAlertById)
	_, err := db.Exec(stmt, id)

	return err
}

func (db *Store) UpdateAlertStatus(id int64, status string) error {
	stmt := queries.Stmt(db.driver, queries.UpdateAlertStatus)
	_, err := db.Exec(stmt, status, id)

	return err
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
