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

type CommitEvent struct {
	CommitStatus map[string]*model.CombinedStatus `json:"commitStatus"`
	StreamingEvent
}
