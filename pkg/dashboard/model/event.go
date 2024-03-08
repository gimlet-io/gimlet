package model

import (
	"encoding/json"
	"fmt"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
)

const StatusNew = "new"
const StatusProcessed = "processed"
const StatusError = "error"

const ArtifactCreatedEvent = "artifact"
const ReleaseRequestedEvent = "release"
const ImageBuildRequestedEvent = "imageBuild"
const RollbackRequestedEvent = "rollback"
const BranchDeletedEvent = "branchDeleted"

type Status int

const (
	Success Status = iota
	Failure
	Pending
)

func (s Status) String() string {
	switch s {
	case Success:
		return "success"
	case Pending:
		return "pending"
	case Failure:
		return "failure"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

type Result struct {
	Manifest        *dx.Manifest
	Artifact        *dx.Artifact
	RollbackRequest *dx.RollbackRequest
	TriggeredBy     string

	Status     Status
	StatusDesc string

	GitopsRef  string
	GitopsRepo string

	TriggeredDeployRequestID string `json:"triggeredDeployRequestID"`
	Log                      string `json:"log"`
}

type Event struct {
	ID         string   `json:"id,omitempty"  meddler:"id"`
	Created    int64    `json:"created,omitempty"  meddler:"created"`
	Type       string   `json:"type,omitempty"  meddler:"type"`
	Blob       string   `json:"blob,omitempty"  meddler:"blob"`
	Status     string   `json:"status"  meddler:"status"`
	StatusDesc string   `json:"statusDesc"  meddler:"status_desc"`
	Results    []Result `json:"results"  meddler:"results,json"`

	// denormalized artifact fields
	Repository   string      `json:"repository,omitempty"  meddler:"repository"`
	Branch       string      `json:"branch,omitempty"  meddler:"branch"`
	Event        dx.GitEvent `json:"event,omitempty"  meddler:"event"`
	SourceBranch string      `json:"sourceBranch,omitempty"  meddler:"source_branch"`
	TargetBranch string      `json:"targetBranch,omitempty"  meddler:"target_branch"`
	Tag          string      `json:"tag,omitempty"  meddler:"tag"`
	SHA          string      `json:"sha"  meddler:"sha"`
	ArtifactID   string      `json:"artifactID"  meddler:"artifact_id"`
}

func ToEvent(artifact dx.Artifact) (*Event, error) {
	artifactStr, err := json.Marshal(artifact)
	if err != nil {
		return nil, fmt.Errorf("cannot serialize artifact: %s", err)
	}

	return &Event{
		Type:         ArtifactCreatedEvent,
		Repository:   artifact.Version.RepositoryName,
		Branch:       artifact.Version.Branch,
		Event:        artifact.Version.Event,
		TargetBranch: artifact.Version.TargetBranch,
		SourceBranch: artifact.Version.SourceBranch,
		Tag:          artifact.Version.Tag,
		Blob:         string(artifactStr),
		SHA:          artifact.Version.SHA,
		ArtifactID:   artifact.ID,
	}, nil
}

func ToArtifact(a *Event) (*dx.Artifact, error) {
	var artifact dx.Artifact
	json.Unmarshal([]byte(a.Blob), &artifact)
	return &artifact, nil
}
