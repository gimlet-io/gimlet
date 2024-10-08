package streaming

import (
	"github.com/gimlet-io/capacitor/pkg/flux"
	"github.com/gimlet-io/gimlet/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
)

const AgentConnectedEventString = "agentConnected"
const AgentDisconnectedEventString = "agentDisconnected"
const EnvsUpdatedEventString = "envsUpdated"
const StaleRepoDataEventString = "staleRepoData"
const GitopsCommitEventString = "gitopsCommit"
const CommitStatusUpdatedEventString = "commitStatusUpdated"
const PodLogsEventString = "podLogs"
const ImageBuildLogEventString = "imageBuildLogEvent"
const FluxStateUpdatedEventString = "fluxStateUpdatedEvent"
const FluxK8sEventsUpdatedEventString = "fluxK8sEventsUpdatedEvent"
const DeploymentDetailsEventString = "deploymentDetailsEvent"
const PodDetailsEventString = "podDetailsEvent"
const AlertPendingEventString = "alertPending"
const AlertFiredEventString = "alertFired"
const AlertResolvedEventString = "alertResolved"
const CommitEventString = "commitEvent"

type StreamingEvent struct {
	Event string `json:"event"`
}

type AgentConnectedEvent struct {
	Agent ConnectedAgent `json:"agent"`
	StreamingEvent
}

type AgentDisconnectedEvent struct {
	Agent ConnectedAgent `json:"agent"`
	StreamingEvent
}

type EnvsUpdatedEvent struct {
	Envs []*api.ConnectedAgent `json:"envs"`
	StreamingEvent
}

type FluxStateUpdatedEvent struct {
	EnvName   string          `json:"envName"`
	FluxState *flux.FluxState `json:"fluxState"`
	StreamingEvent
}

type FluxK8sEventsUpdatedEvent struct {
	EnvName    string        `json:"envName"`
	FluxEvents []*flux.Event `json:"fluxEvents"`
	StreamingEvent
}

type DeploymentDetailsEvent struct {
	Deployment string `json:"deployment"`
	Details    string `json:"details"`
	StreamingEvent
}

type PodDetailsEvent struct {
	Pod     string `json:"pod"`
	Details string `json:"details"`
	StreamingEvent
}

type StaleRepoDataEvent struct {
	Repo string `json:"repo"`
	StreamingEvent
}

type GitopsEvent struct {
	GitopsCommit interface{} `json:"gitopsCommit"`
	StreamingEvent
}

type ImageBuildLogEvent struct {
	BuildId string `json:"buildId"`
	Status  string `json:"status"`
	LogLine string `json:"logLine,omitempty"`
	StreamingEvent
}

type CommitStatusUpdatedEvent struct {
	CommitStatus  *model.CombinedStatus `json:"commitStatus"`
	Owner         string                `json:"owner"`
	Sha           string                `json:"sha"`
	RepoName      string                `json:"repo"`
	DeployTargets []*api.DeployTarget   `json:"deployTargets"`
	StreamingEvent
}

type PodLogsEvent struct {
	Timestamp  string `json:"timestamp"`
	Container  string `json:"container"`
	Message    string `json:"message"`
	Pod        string `json:"pod"`
	Deployment string `json:"deployment"`
	StreamingEvent
}

type AlertEvent struct {
	Alert *api.Alert `json:"alert"`
	StreamingEvent
}

type CommitEvent struct {
	CommitEvent *api.CommitEvent `json:"commitEvent"`
	StreamingEvent
}
