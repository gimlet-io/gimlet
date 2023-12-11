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

func (db *Store) AlertsInWeek(since time.Time, until time.Time) ([]*model.Alert, error) {
	query := sql.Stmt(db.driver, sql.SelectAlerts)
	data := []*model.Alert{}
	err := meddler.QueryAll(db, &data, query, xx)
	return data, err
}

func (db *Store) AlertsBetweenPreviousTwoWeeks() ([]*model.Alert, error) {
	query := sql.Stmt(db.driver, sql.SelectAlertsInterval)
	data := []*model.Alert{}
	weekAgo := time.Now().Add(-7 * time.Hour * 24).Unix()
	twoWeeksAgo := time.Now().Add(-14 * time.Hour * 24).Unix()
	err := meddler.QueryAll(db, &data, query, weekAgo, twoWeeksAgo)
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

func (db *Store) UpdateAlertState(alert *model.Alert) error {
	var query string

	var stamp int64
	if alert.Status == model.FIRING {
		query = sql.UpdateAlertStatusFired
		stamp = alert.FiredAt
	} else if alert.Status == model.RESOLVED {
		query = sql.UpdateAlertStatusResolved
		stamp = alert.ResolvedAt
	} else {
		return fmt.Errorf("invalid status provided")
	}

	stmt := sql.Stmt(db.driver, query)
	_, err := db.Exec(stmt, alert.Status, stamp, alert.ID)
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
