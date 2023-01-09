package model

type Pod struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Name       string `json:"name,omitempty"  meddler:"name"`
	Namespace  string `json:"deploymentNamespace,omitempty"  meddler:"deployment_namespace"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"statusDesc"`
}
