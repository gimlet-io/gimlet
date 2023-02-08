package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) PendingPods() ([]*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPendingPods)
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

func (db *Store) DeletePod(name string) error {
	stmt := queries.Stmt(db.driver, queries.DeletePodByName)
	_, err := db.Exec(stmt, name)

	return err
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

	storedPod.DeploymentName = pod.DeploymentName
	storedPod.Status = pod.Status
	storedPod.StatusDesc = pod.StatusDesc
	storedPod.AlertState = pod.AlertState
	storedPod.AlertStateTimestamp = pod.AlertStateTimestamp

	return meddler.Update(db, "pods", storedPod)
}
