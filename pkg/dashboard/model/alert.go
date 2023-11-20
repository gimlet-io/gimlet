package model

import "time"

const POD_ALERT = "pod"

const PENDING = "Pending"
const FIRING = "Firing"
const RESOLVED = "Resolved"

type Alert struct {
	ID             int64  `json:"-"  meddler:"id,pk"`
	Type           string `json:"type,omitempty"  meddler:"type"`
	ObjectName     string `json:"objectName,omitempty"  meddler:"name"` // TODO rename this to object_name in db
	DeploymentName string `json:"deploymentName,omitempty"  meddler:"deployment_name"`
	Status         string `json:"status,omitempty"  meddler:"status"`
	ImChannelId    string `json:"imChannelId,omitempty"  meddler:"im_channel_id"`
	DeploymentUrl  string `json:"deploymentUrl,omitempty"  meddler:"deployment_url"`
	PendingAt      int64  `json:"pendingAt,omitempty"  meddler:"pending_at"`
	FiredAt        int64  `json:"firedAt,omitempty"  meddler:"fired_at"`
	ResolvedAt     int64  `json:"resolvedAt,omitempty"  meddler:"resolved_at"`
}

func (a *Alert) SetFiring() {
	a.Status = FIRING
	a.FiredAt = time.Now().Unix()
}

func (a *Alert) SetResolved() {
	a.Status = RESOLVED
	a.ResolvedAt = time.Now().Unix()
}
