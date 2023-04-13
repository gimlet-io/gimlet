package model

const ALERT_STATE_FIRING = "Firing"
const ALERT_STATE_PENDING = "Pending"
const ALERT_STATE_RESOLVED = "Resolved"

const ALERT_OBJECT_TYPE_POD = "pod"
const ALERT_OBJECT_TYPE_EVENT = "event"

type Alert struct {
	ID int64 `json:"-"  meddler:"id,pk"`

	// TODO rename this to ObjectType
	Type string `json:"type,omitempty"  meddler:"type"`
	// TODO rename this to ObjectName
	Name             string `json:"name,omitempty"  meddler:"name"`
	DeploymentName   string `json:"deploymentName,omitempty"  meddler:"deployment_name"`
	ObjectStatus     string `json:"objectStatus,omitempty"  meddler:"object_status"`
	ObjectStatusDesc string `json:"objectStatusDesc,omitempty"  meddler:"object_status_desc"`

	Status string `json:"status,omitempty"  meddler:"status"`
	// StatusDesc      string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	LastStateChange int64 `json:"lastStateChange,omitempty"  meddler:"last_state_change"`
}

func (a *Alert) IsFiring() bool {
	return a.Status == ALERT_STATE_FIRING
}

func (a *Alert) EvaluateThreshold() bool {
	return false
}
