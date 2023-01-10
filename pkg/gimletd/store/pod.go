package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/gimletd/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Pods() ([]*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectAllPods)
	data := []*model.Pod{}
	err := meddler.QueryAll(db, &data, stmt)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) Pod(name string) (*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPodByName)
	pod := new(model.Pod)
	err := meddler.QueryRow(db, pod, stmt, name)

	return pod, err
}

func (db *Store) SaveOrUpdatePod(pod *model.Pod) error {
	storedPod, err := db.Pod(pod.Name)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return meddler.Insert(db, "pods", pod)
		default:
			return err
		}
	}

	storedPod.Status = pod.Status
	storedPod.StatusDesc = pod.StatusDesc
	storedPod.AlertState = pod.AlertState
	storedPod.AlertStateTimestamp = pod.AlertStateTimestamp

	return meddler.Update(db, "pods", storedPod)
}
