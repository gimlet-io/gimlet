package dx

// Release contains all metadata about a release event
type Release struct {
	App string `json:"app"`
	Env string `json:"env"`

	ArtifactID  string `json:"artifactId"`
	TriggeredBy string `json:"triggeredBy"`

	Version *Version `json:"version"`

	GitopsRef  string `json:"gitopsRef"`
	GitopsRepo string `json:"gitopsRepo"`
	Created    int64  `json:"created,omitempty"`

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

//GitopsStatus holds the gitops references that were created based on an event
type GitopsStatus struct {
	Hash       string `json:"hash,omitempty"`
	Status     string `json:"status,omitempty"`
	StatusDesc string `json:"statusDesc,omitempty"`
}

type Result struct {
	App        string `json:"app,omitempty"`
	Hash       string `json:"hash,omitempty"`
	Status     string `json:"status,omitempty"`
	StatusDesc string `json:"statusDesc,omitempty"`
}

type ReleaseStatus struct {
	Status       string         `json:"status"`
	StatusDesc   string         `json:"statusDesc"`
	GitopsHashes []GitopsStatus `json:"gitopsHashes"`
	Results      []Result       `json:"results"`
}
