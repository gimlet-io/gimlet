package model

const (
	POD_PENDING    = "Pending"
	POD_RUNNING    = "Running"
	POD_TERMINATED = "Terminated"
)

type Pod struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Name       string `json:"name,omitempty"  meddler:"name"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"status_desc"`
}
