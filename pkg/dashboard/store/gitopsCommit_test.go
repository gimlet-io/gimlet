package store

import (
	"testing"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/stretchr/testify/assert"
)

func TestGitopsCommitCRUD(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	gitopsCommit := &model.GitopsCommit{
		Sha: "sha",
	}

	_, err := s.SaveOrUpdateGitopsCommit(gitopsCommit)
	assert.Nil(t, err)

	gitopsCommit.Status = "aStatus"
	_, err = s.SaveOrUpdateGitopsCommit(gitopsCommit)
	assert.Nil(t, err)

	savedGitopsCommit, err := s.GitopsCommit("sha")
	assert.Nil(t, err)
	assert.Equal(t, "aStatus", savedGitopsCommit.Status)
}
