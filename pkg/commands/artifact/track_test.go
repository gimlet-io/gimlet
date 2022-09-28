package artifact

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"gotest.tools/assert"
)

func Test_processArifactStatusNew(t *testing.T) {
	testReleaseStatus := dx.ReleaseStatus{
		Status:       "new",
		StatusDesc:   "",
		GitopsHashes: []dx.GitopsStatus{},
		Results:      []dx.Result{},
	}

	statusError, everySucceeded, hasFailed := processArtifactStatus(testReleaseStatus)

	assert.Equal(t, statusError, false)
	assert.Equal(t, everySucceeded, false)
	assert.Equal(t, hasFailed, false)
}

func Test_processArifactStatusError(t *testing.T) {
	testReleaseStatus := dx.ReleaseStatus{
		Status:       "error",
		StatusDesc:   "",
		GitopsHashes: []dx.GitopsStatus{},
		Results:      []dx.Result{},
	}

	statusError, everySucceeded, hasFailed := processArtifactStatus(testReleaseStatus)

	assert.Equal(t, statusError, true)
	assert.Equal(t, everySucceeded, false)
	assert.Equal(t, hasFailed, false)
}

func Test_processArifactStatusGitopsCommitFailed(t *testing.T) {
	testReleaseStatus := dx.ReleaseStatus{
		Status:     "processed",
		StatusDesc: "",
		GitopsHashes: []dx.GitopsStatus{
			{
				Hash:       "abc123",
				Status:     "ReconciliationFailed",
				StatusDesc: "test description one",
			},
			{
				Hash:       "abc456",
				Status:     "ReconciliationSucceeded",
				StatusDesc: "test description two",
			},
		},
		Results: []dx.Result{
			{
				Hash:               "abc123",
				GitopsCommitStatus: "ReconciliationFailed",
				StatusDesc:         "test description one",
			},
			{
				Hash:               "abc456",
				GitopsCommitStatus: "ReconciliationSucceeded",
				StatusDesc:         "test description two",
			},
		},
	}

	statusError, everySucceeded, hasFailed := processArtifactStatus(testReleaseStatus)

	assert.Equal(t, statusError, false)
	assert.Equal(t, everySucceeded, false)
	assert.Equal(t, hasFailed, true)
}

func Test_processArifactStatusGitopsCommitsSucceeded(t *testing.T) {
	testReleaseStatus := dx.ReleaseStatus{
		Status:     "processed",
		StatusDesc: "",
		GitopsHashes: []dx.GitopsStatus{
			{
				Hash:       "abc123",
				Status:     "ReconciliationSucceeded",
				StatusDesc: "test description one",
			},
			{
				Hash:       "abc456",
				Status:     "ReconciliationSucceeded",
				StatusDesc: "test description two",
			},
		},
		Results: []dx.Result{
			{
				Hash:               "abc123",
				GitopsCommitStatus: "ReconciliationSucceeded",
				StatusDesc:         "test description one",
			},
			{
				Hash:               "abc456",
				GitopsCommitStatus: "ReconciliationSucceeded",
				StatusDesc:         "test description two",
			},
		},
	}

	statusError, everySucceeded, hasFailed := processArtifactStatus(testReleaseStatus)

	assert.Equal(t, statusError, false)
	assert.Equal(t, everySucceeded, true)
	assert.Equal(t, hasFailed, false)
}
