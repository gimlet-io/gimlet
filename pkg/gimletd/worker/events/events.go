package events

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
)

type Status int

const (
	Success Status = iota
	Failure
)

type RollbackEvent struct {
	RollbackRequest *dx.RollbackRequest

	Status     Status
	StatusDesc string

	GitopsRefs []string
	GitopsRepo string
}

type DeleteEvent struct {
	Env         string
	App         string
	TriggeredBy string

	Status     Status
	StatusDesc string

	GitopsRef          string
	GitopsRepo         string
	BranchDeletedEvent BranchDeletedEvent
}

// BranchDeletedEvent contains all metadata about the deleted branch
type BranchDeletedEvent struct {
	Manifests []*dx.Manifest
	Branch    string
	Repo      string
}
