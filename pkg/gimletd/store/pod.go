package store

import (
	"database/sql"
	"fmt"

	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/gimletd/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) Pods() ([]*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectGitopsCommits)
	data := []*model.Pod{}
	err := meddler.QueryAll(db, &data, stmt)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) Pod(deployment string) (*model.Pod, error) {
	stmt := queries.Stmt(db.driver, queries.SelectPodByDeployment)
	pod := new(model.Pod)
	err := meddler.QueryRow(db, pod, stmt, deployment)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return pod, err
}

func (db *Store) SaveOrUpdatePod(pod *model.Pod) error {
	stmt := queries.Stmt(db.driver, queries.SelectPodByDeployment)
	savedPod := new(model.Pod)
	err := meddler.QueryRow(db, savedPod, stmt, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	if err == sql.ErrNoRows {
		return meddler.Insert(db, "pods", pod)
	} else if err != nil {
		return err
	}

	savedPod.Status = pod.Status
	savedPod.StatusDesc = pod.StatusDesc

	return meddler.Update(db, "pods", savedPod)
}
