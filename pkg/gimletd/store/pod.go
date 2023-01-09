package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/gimletd/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Pod(deployment string) (*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPodByDeployment)
	pod := new(model.Pod)
	err := meddler.QueryRow(db, pod, stmt, deployment)

	return pod, err
}

func (db *Store) SaveOrUpdatePod(pod *model.Pod) error {
	storedPod, err := db.Pod(pod.Deployment)

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

	return meddler.Update(db, "pods", storedPod)
}
