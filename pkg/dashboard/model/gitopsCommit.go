package model

const Progressing = "Progressing"
const ReconciliationSucceeded = "ReconciliationSucceeded"
const ValidationFailed = "ValidationFailed"
const ReconciliationFailed = "ReconciliationFailed"
const HealthCheckFailed = "HealthCheckFailed"
const NotReconciled = "NotReconciled"

type GitopsCommit struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Sha        string `json:"sha,omitempty"  meddler:"sha"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"status_desc"`
	Created    int64  `json:"created,omitempty"  meddler:"created"`
	Env        string `json:"env,omitempty"  meddler:"env"`
}
