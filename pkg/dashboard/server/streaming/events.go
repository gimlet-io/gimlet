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
const IrregularPodsString = "irregularPods"
const PodLogsEventString = "podLogs"
const KubernetesEventsString = "kubernetesEvents"

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

type StaleRepoDataEvent struct {
	Repo string `json:"repo"`
	StreamingEvent
}

type GitopsEvent struct {
	GitopsCommit interface{} `json:"gitopsCommit"`
	StreamingEvent
}

type CommitStatusUpdatedEvent struct {
	CommitStatus  *model.CombinedStatus `json:"commitStatus"`
	Owner         string                `json:"owner"`
	Sha           string                `json:"sha"`
	RepoName      string                `json:"repo"`
	DeployTargets []*model.DeployTarget `json:"deployTargets"`
	StreamingEvent
}

type IrregularPodsEvent struct {
	IrregularPods []api.Alert `json:"irregularPods"`
	StreamingEvent
}

type PodLogsEvent struct {
	PodLogs string `json:"podLogs"`
	Pod     string `json:"pod"`
	StreamingEvent
}

type KubernetesEventsEvent struct {
	KubernetesEvents []api.Alert `json:"kubernetesEvents"`
	StreamingEvent
}
