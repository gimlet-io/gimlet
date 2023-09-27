package model

const POD_ALERT = "pod"

const PENDING = "Pending"
const FIRING = "Firing"
const RESOLVED = "Resolved"

type Alert struct {
	ID             int64  `json:"-"  meddler:"id,pk"`
	Type           string `json:"type,omitempty"  meddler:"type"`
	ObjectName     string `json:"objectName,omitempty"  meddler:"name"` // TODO rename this to object_name in db
	DeploymentName string `json:"deploymentName,omitempty"  meddler:"deployment_name"`
	Status         string `json:"status,omitempty"  meddler:"status"`
	// LastStateChange int64  `json:"lastStateChange,omitempty"  meddler:"last_state_change"` // TODO remove this from db
	// Count           int32  `json:"count"  meddler:"count"`                                 // TODO remove this from db
	CreatedAt  int64 `json:"createdAt,omitempty"  meddler:"created_at"`
	ReachedAt  int64 `json:"reachedAt,omitempty"  meddler:"reached_at"`
	ResolvedAt int64 `json:"resolvedAt,omitempty"  meddler:"resolved_at"`
}
