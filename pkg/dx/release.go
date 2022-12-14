package dx

import (
	"strings"
)

const Progressing = "Progressing"
const ReconciliationSucceeded = "ReconciliationSucceeded"
const ValidationFailed = "ValidationFailed"
const ReconciliationFailed = "ReconciliationFailed"
const HealthCheckFailed = "HealthCheckFailed"
const NotReconciled = "NotReconciled"

// Release contains all metadata about a release event
type Release struct {
	App string `json:"app"`
	Env string `json:"env"`

	ArtifactID  string `json:"artifactId"`
	TriggeredBy string `json:"triggeredBy"`

	Version *Version `json:"version"`

	GitopsRef              string `json:"gitopsRef"`
	GitopsRepo             string `json:"gitopsRepo"`
	GitopsCommitStatus     string `json:"gitopsCommitStatus"`
	GitopsCommitStatusDesc string `json:"gitopsCommitStatusDesc"`
	GitopsCommitCreated    int64  `json:"gitopsCommitCreated,omitempty"`
	Created                int64  `json:"created,omitempty"`

	RolledBack bool `json:"rolledBack,omitempty"`
}

// ReleaseRequest contains all metadata about the release intent
type ReleaseRequest struct {
	Env         string `json:"env"`
	App         string `json:"app,omitempty"`
	ArtifactID  string `json:"artifactId"`
	TriggeredBy string `json:"triggeredBy"`
}

// RollbackRequest contains all metadata about the rollback intent
type RollbackRequest struct {
	Env         string `json:"env"`
	App         string `json:"app"`
	TargetSHA   string `json:"targetSHA"`
	TriggeredBy string `json:"triggeredBy"`
}

// Result of processing the Gimlet environment file or files from artifact
type Result struct {
	App  string `json:"app,omitempty"`
	Hash string `json:"hash,omitempty"`
	// Status of processing the Gimlet environment file from artifact
	Status                 string `json:"status,omitempty"`
	GitopsCommitStatus     string `json:"gitopsCommitStatus,omitempty"`
	GitopsCommitStatusDesc string `json:"gitopsCommitStatusDesc,omitempty"`
	Env                    string `json:"env,omitempty"`
	StatusDesc             string `json:"statusDesc,omitempty"`
}

type ReleaseStatus struct {
	Status     string   `json:"status"`
	StatusDesc string   `json:"statusDesc"`
	Results    []Result `json:"results"`
}

func (rs *ReleaseStatus) ExtractGitopsEndState() (bool, bool) {
	var artifactResultCount int
	var failedCount int
	var succeededCount int
	var allCommitsApplied bool
	var gitopsCommitsHaveFailed bool

	artifactResultCount = len(rs.Results)

	for _, result := range rs.Results {
		if strings.Contains(result.GitopsCommitStatus, "Failed") || result.Status == "failure" {
			failedCount++
		} else if result.GitopsCommitStatus == ReconciliationSucceeded {
			succeededCount++
		}
	}

	if succeededCount == artifactResultCount {
		allCommitsApplied = true
	}

	if failedCount > 0 {
		gitopsCommitsHaveFailed = true
	}

	return allCommitsApplied, gitopsCommitsHaveFailed
}
