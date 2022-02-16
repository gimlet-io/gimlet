package store

import (
	databaseSql "database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store/sql"
	"github.com/russross/meddler"
	"github.com/sirupsen/logrus"
)

// CreateCommit stores a new commit in the database
func (db *Store) CreateCommit(commit *model.Commit) error {
	return meddler.Insert(db, "commits", commit)
}

// SaveCommits stores new commits in the database, or updates them
func (db *Store) SaveCommits(repo string, commits []*model.Commit) error {
	hashes := []string{}
	for _, c := range commits {
		hashes = append(hashes, c.SHA)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	existingSHAs, err := db.commitShasByRepoAndSHA(tx, repo, hashes)
	if err != nil {
		return err
	}

	shaMap := map[string]*model.Commit{}
	for _, c := range existingSHAs {
		shaMap[c.SHA] = c
	}

	commitsToInsert := []*model.Commit{}
	for _, c := range commits {
		if _, exists := shaMap[c.SHA]; !exists {
			commitsToInsert = append(commitsToInsert, c)
		}
	}

	if len(commitsToInsert) != 0 {
		valueStrings := make([]string, 0, len(commitsToInsert))
		valueArgs := make([]interface{}, 0, len(commitsToInsert)*9)
		for idx, c := range commitsToInsert {
			valueStrings = append(valueStrings,
				fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
					idx*9+1, idx*9+2, idx*9+3, idx*9+4, idx*9+5, idx*9+6, idx*9+7, idx*9+8, idx*9+9))
			valueArgs = append(valueArgs, repo)
			valueArgs = append(valueArgs, c.SHA)
			valueArgs = append(valueArgs, c.URL)
			valueArgs = append(valueArgs, c.Author)
			valueArgs = append(valueArgs, c.AuthorPic)
			valueArgs = append(valueArgs, c.Message)
			valueArgs = append(valueArgs, c.Created)
			valueArgs = append(valueArgs, "[]")
			valueArgs = append(valueArgs, "{}")
		}
		stmt := fmt.Sprintf("INSERT INTO commits (repo, sha, url, author, author_pic, message, created, tags, status) VALUES %s", strings.Join(valueStrings, ","))
		_, err = tx.Exec(stmt, valueArgs...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// SaveTagsOnCommits updates tags on commits
func (db *Store) SaveTagsOnCommits(repo string, tags map[string][]string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for sha, tags := range tags {
		tagsJson, err := json.Marshal(tags)
		if err != nil {
			tx.Rollback()
			return err
		}
		stmt := "UPDATE commits set tags = $1 where repo = $2 and sha = $3"
		_, err = tx.Exec(stmt, tagsJson, repo, sha)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// SaveStatusesOnCommits updates statuses on commits
func (db *Store) SaveStatusesOnCommits(repo string, statuses map[string]*model.CombinedStatus) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for sha, status := range statuses {
		statusJson, err := json.Marshal(status)
		if err != nil {
			tx.Rollback()
			return err
		}
		logrus.Infof("Saving status on commit %s %s: %s", sha, repo, statusJson)
		stmt := "UPDATE commits set status = $1 where repo = $2 and sha = $3"
		_, err = tx.Exec(stmt, statusJson, repo, sha)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// Commits gets the most recent 20 commits from a repo
func (db *Store) Commits(repo string) ([]*model.Commit, error) {
	stmt := sql.Stmt(db.driver, sql.SelectCommitsByRepo)
	data := []*model.Commit{}
	err := meddler.QueryAll(db, &data, stmt, repo)
	return data, err
}

func (db *Store) commitShasByRepoAndSHA(tx *databaseSql.Tx, repo string, hashes []string) ([]*model.Commit, error) {
	if len(hashes) == 0 {
		return []*model.Commit{}, nil
	}

	filters := []string{}
	args := []interface{}{repo}

	for _, s := range hashes {
		filters = append(filters, fmt.Sprintf("$%d", len(filters)+2))
		args = append(args, s)
	}

	stmt := fmt.Sprintf(`
select sha from commits where repo=$1
and sha in (%s)
`, strings.Join(filters, ","))

	data := []*model.Commit{}
	err := meddler.QueryAll(tx, &data, stmt, args...)

	return data, err
}

func (db *Store) CommitsByRepoAndSHA(repo string, hashes []string) ([]*model.Commit, error) {
	if len(hashes) == 0 {
		return []*model.Commit{}, nil
	}

	filters := []string{}
	args := []interface{}{repo}

	for _, s := range hashes {
		filters = append(filters, fmt.Sprintf("$%d", len(filters)+2))
		args = append(args, s)
	}

	stmt := fmt.Sprintf(`
select sha, url, author, author_pic, tags, status, message, created from commits where repo=$1 
and sha in (%s)
`, strings.Join(filters, ","))

	data := []*model.Commit{}
	err := meddler.QueryAll(db, &data, stmt, args...)

	return data, err
}
