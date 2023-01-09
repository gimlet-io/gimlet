package model

type Pod struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Deployment string `json:"deployment,omitempty"  meddler:"deployment"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"status_desc"`
}
