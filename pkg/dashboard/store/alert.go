package store

import (
	"time"

	sys_sql "database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Alerts() ([]*model.Alert, error) {
	query := sql.Stmt(db.driver, sql.SelectAlerts)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, query)
	return data, err
}

func (db *Store) AlertsByState(status string) ([]*model.Alert, error) {
	stmt := sql.Stmt(db.driver, sql.SelectAlertsByState)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, stmt, status)

	if err == sys_sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) UpdateAlertState(id int64, status string) error {
	stmt := sql.Stmt(db.driver, sql.UpdateAlertStatus)
	_, err := db.Exec(stmt, status, time.Now().Unix(), id)
	return err
}

func (db *Store) CreateAlert(alert *model.Alert) (*model.Alert, error) {
	return alert, meddler.Insert(db, "alerts", alert)
}

func (db *Store) RelatedAlerts(name string) ([]*model.Alert, error) {
	stmt := sql.Stmt(db.driver, sql.SelectAlertsByName)
	alerts := []*model.Alert{}
	err := meddler.QueryAll(db, &alerts, stmt, name)

	return alerts, err
}
