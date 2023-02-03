package model

type KubernetesEvent struct {
	ID                  int64  `json:"-"  meddler:"id,pk"`
	FirstTimestamp      int64  `json:"firstTimestamp"  meddler:"first_timestamp"`
	Count               int32  `json:"count"  meddler:"count"`
	Name                string `json:"name"  meddler:"name"`
	Status              string `json:"status"  meddler:"status"`
	StatusDesc          string `json:"statusDesc"  meddler:"status_desc"`
	AlertState          string `json:"alertState,omitempty"  meddler:"alert_state"`
	AlertStateTimestamp int64  `json:"alertStateTimestamp,omitempty"  meddler:"alert_state_timestamp"`
}
