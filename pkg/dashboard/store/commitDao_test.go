package store

import (
	"testing"
	"time"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestCommitCRUD(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	aTime := time.Now()
	commit := model.Commit{
		Repo:      "aRepo",
		SHA:       "asha",
		URL:       "aUrl",
		Author:    "anAuthor",
		AuthorPic: "anAuthorPic",
		Created:   aTime.Unix(),
		Tags:      []string{"aTag", "another"},
	}

	err := s.CreateCommit(&commit)
	assert.Nil(t, err)

	commits, err := s.Commits("aRepo")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(commits))
	assert.Equal(t, aTime.Unix(), commits[0].Created)
}

func TestBulkCommitCreate(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	commits := []*model.Commit{
		{
			Repo: "aRepo",
			SHA:  "aSha",
		},
	}

	err := s.SaveCommits("aRepo", commits)
	assert.Nil(t, err)

	commits, err = s.Commits("aRepo")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(commits))
	assert.Equal(t, "aSha", commits[0].SHA)
}

func TestBulkQuery(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	commits := []*model.Commit{
		{
			Repo: "aRepo",
			SHA:  "aSha",
		},
		{
			Repo: "aRepo",
			SHA:  "anotherSha",
		},
	}

	err := s.SaveCommits("aRepo", commits)
	assert.Nil(t, err)

	commits, err = s.CommitsByRepoAndSHA("aRepo", []string{"aSha", "anotherSha"})
	assert.Nil(t, err)
	assert.True(t, len(commits) == 2)
}
