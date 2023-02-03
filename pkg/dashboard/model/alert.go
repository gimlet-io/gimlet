package model

type Alert struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Type       string `json:"type,omitempty"  meddler:"type"`
	Name       string `json:"name,omitempty"  meddler:"name"`
	Env        string `json:"env,omitempty"  meddler:"env"`
	Repo       string `json:"repo,omitempty"  meddler:"repo"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	Fired      int64  `json:"fired,omitempty"  meddler:"fired"`
}
