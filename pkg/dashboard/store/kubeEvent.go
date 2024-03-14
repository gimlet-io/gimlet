package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) SaveOrUpdateKubeEvent(event *model.KubeEvent) error {
	storedEvent, err := db.KubeEvent(event.Name)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return meddler.Insert(db, "kube_events", event)
		default:
			return err
		}
	}

	storedEvent.Status = event.Status
	storedEvent.StatusDesc = event.StatusDesc

	return meddler.Update(db, "kube_events", storedEvent)
}

func (db *Store) KubeEvent(name string) (*model.KubeEvent, error) {
	stmt := queries.Stmt(db.driver, queries.SelectKubeEventByName)
	event := new(model.KubeEvent)
	err := meddler.QueryRow(db, event, stmt, name)

	return event, err
}

func (db *Store) DeleteEvent(name string) error {
	stmt := queries.Stmt(db.driver, queries.DeleteKubeEventByName)
	_, err := db.Exec(stmt, name)

	return err
}
