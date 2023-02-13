package model

type Alert struct {
	ID             int64  `json:"-"  meddler:"id,pk"`
	Type           string `json:"type,omitempty"  meddler:"type"`
	Name           string `json:"name,omitempty"  meddler:"name"`
	DeploymentName string `json:"deploymentName,omitempty"  meddler:"deployment_name"`
	Status         string `json:"status,omitempty"  meddler:"status"`
	StatusDesc     string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	Fired          int64  `json:"fired,omitempty"  meddler:"fired"`
	Resolved       int64  `json:"resolved,omitempty"  meddler:"resolved"`
	Count          int32  `json:"count"  meddler:"count"`
}
