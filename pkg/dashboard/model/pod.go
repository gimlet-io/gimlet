package model

type Pod struct {
	ID                  int64  `json:"-"  meddler:"id,pk"`
	Name                string `json:"name,omitempty"  meddler:"name"`
	Status              string `json:"status,omitempty"  meddler:"status"`
	StatusDesc          string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	AlertState          string `json:"alertState,omitempty"  meddler:"alert_state"`
	AlertStateTimestamp int64  `json:"alertStateTimestamp,omitempty"  meddler:"alert_state_timestamp"`
}
