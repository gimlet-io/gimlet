package store

import (
	"database/sql"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	queries "github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
)

func (db *Store) GitopsCommit(sha string) (*model.GitopsCommit, error) {
	stmt := queries.Stmt(db.driver, queries.SelectGitopsCommitBySha)
	gitopsCommit := new(model.GitopsCommit)
	err := meddler.QueryRow(db, gitopsCommit, stmt, sha)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return gitopsCommit, err
}

func (db *Store) GitopsCommits() ([]*model.GitopsCommit, error) {
	stmt := queries.Stmt(db.driver, queries.SelectGitopsCommits)
	data := []*model.GitopsCommit{}
	err := meddler.QueryAll(db, &data, stmt)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return data, err
}

func (db *Store) SaveOrUpdateGitopsCommit(gitopsCommit *model.GitopsCommit) error {
	stmt := queries.Stmt(db.driver, queries.SelectGitopsCommitBySha)
	savedGitopsCommit := new(model.GitopsCommit)
	err := meddler.QueryRow(db, savedGitopsCommit, stmt, gitopsCommit.Sha)
	if err == sql.ErrNoRows {
		return meddler.Insert(db, "gitops_commits", gitopsCommit)
	} else if err != nil {
		return err
	}

	savedGitopsCommit.Status = gitopsCommit.Status
	savedGitopsCommit.StatusDesc = gitopsCommit.StatusDesc
	savedGitopsCommit.Created = gitopsCommit.Created
	savedGitopsCommit.Env = gitopsCommit.Env
	return meddler.Update(db, "gitops_commits", savedGitopsCommit)
}
