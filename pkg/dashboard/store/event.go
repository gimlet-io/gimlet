package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) SaveOrUpdateEvent(event *model.Event) error {
	storedEvent, err := db.Event(event.Name)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return meddler.Insert(db, "kubernetes_events", event)
		default:
			return err
		}
	}

	storedEvent.Status = event.Status
	storedEvent.StatusDesc = event.StatusDesc

	return meddler.Update(db, "kubernetes_events", storedEvent)
}

func (db *Store) Event(name string) (*model.Event, error) {
	stmt := queries.Stmt(db.driver, queries.SelectEventByName)
	alert := new(model.Event)
	err := meddler.QueryRow(db, alert, stmt, name)

	return alert, err
}

func (db *Store) DeleteEvent(name string) error {
	stmt := queries.Stmt(db.driver, queries.DeleteEventByName)
	_, err := db.Exec(stmt, name)

	return err
}
