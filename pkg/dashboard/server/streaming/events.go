package streaming

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
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
	EnvName   string         `json:"envName"`
	FluxState *api.FluxState `json:"fluxState"`
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
	Timestamp string `json:"timestamp"`
	Container string `json:"container"`
	Message   string `json:"message"`
	Pod       string `json:"pod"`
	StreamingEvent
}
