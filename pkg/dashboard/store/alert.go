package store

import (
	"fmt"
	"time"

	sys_sql "database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Alerts() ([]*model.Alert, error) {
	query := sql.Stmt(db.driver, sql.SelectAlerts)
	data := []*model.Alert{}
	twentyFourHoursAgo := time.Now().Add(-1 * time.Hour * 24).Unix()
	err := meddler.QueryAll(db, &data, query, twentyFourHoursAgo)
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
	var query string

	if status == model.FIRING {
		query = sql.UpdateAlertStatusFired
	} else if status == model.RESOLVED {
		query = sql.UpdateAlertStatusResolved
	} else {
		return fmt.Errorf("invalid status provided")
	}

	stmt := sql.Stmt(db.driver, query)
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

func (db *Store) AlertsByDeployment(name string) ([]*model.Alert, error) {
	stmt := sql.Stmt(db.driver, sql.SelectAlertsByDeploymentName)
	alerts := []*model.Alert{}
	err := meddler.QueryAll(db, &alerts, stmt, name)

	return alerts, err
}
