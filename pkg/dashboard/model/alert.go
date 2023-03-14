package model

type Alert struct {
	ID              int64  `json:"-"  meddler:"id,pk"`
	Type            string `json:"type,omitempty"  meddler:"type"`
	Name            string `json:"name,omitempty"  meddler:"name"`
	DeploymentName  string `json:"deploymentName,omitempty"  meddler:"deployment_name"`
	Status          string `json:"status,omitempty"  meddler:"status"`
	StatusDesc      string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	LastStateChange int64  `json:"lastStateChange,omitempty"  meddler:"last_state_change"`
	Count           int32  `json:"count"  meddler:"count"`
}
