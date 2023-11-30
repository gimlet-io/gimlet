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
	Tenant      string `json:"tenant"`
	ArtifactID  string `json:"artifactId"`
	TriggeredBy string `json:"triggeredBy"`
}

// ImageBuildRequest contains all metadata to be able to build an image
type ImageBuildRequest struct {
	Env         string `json:"env"`
	App         string `json:"app"`
	Sha         string `json:"sha"`
	TriggeredBy string `json:"triggeredBy"`
	ArtifactID  string `json:"artifactID"`
	Image       string `json:"image"`
	Tag         string `json:"tag"`
	SourcePath  string `json:"sourcePath"`
	AppSource   string `json:"source"`
	Dockerfile  string `json:"dockerfile"`
}

// RollbackRequest contains all metadata about the rollback intent
type RollbackRequest struct {
	Env         string `json:"env"`
	App         string `json:"app"`
	TargetSHA   string `json:"targetSHA"`
	TriggeredBy string `json:"triggeredBy"`
}

// Result of the Gimlet environment manifest processing
type Result struct {
	App  string `json:"app,omitempty"`
	Hash string `json:"hash,omitempty"`
	// Status of the Gimlet environment manifest processing
	Status string `json:"status,omitempty"`
	// GitopsCommitStatus shows the status of the gitops commit of the Gimlet environment manifest processing
	// While manifests are processed succesfully, and gitops commits are written, Flux may fail to apply them
	// This field holds the Flux results
	GitopsCommitStatus       string `json:"gitopsCommitStatus,omitempty"`
	GitopsCommitStatusDesc   string `json:"gitopsCommitStatusDesc,omitempty"`
	Env                      string `json:"env,omitempty"`
	StatusDesc               string `json:"statusDesc,omitempty"`
	TriggeredDeployRequestID string `json:"triggeredDeployRequestID,omitempty"`
}

// ReleaseStatus is the result of an artifact shipping or an on-demand deploy
type ReleaseStatus struct {
	Type string `json:"type"`

	// Status of the artifact processing or an on-demand deploy event's processing
	Status string `json:"status"`

	// StatusDesc is the longer format of the processing Status
	StatusDesc string `json:"statusDesc"`

	// An artifact or an on-demand deploy tipycally holds multiple Gimlet environment manifest configuration
	// Results is the result of the processing of each manifest, one manifest can fail, while others may succeed
	Results []Result `json:"results"`
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
