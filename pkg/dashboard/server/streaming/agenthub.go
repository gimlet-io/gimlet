package streaming

import (
	"encoding/json"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
)

// ConnectedAgent represents a connected k8s cluster
type ConnectedAgent struct {
	Name         string       `json:"name"`
	Namespace    string       `json:"namespace"`
	EventChannel chan []byte  `json:"-"`
	Stacks       []*api.Stack `json:"-"`
}

// AgentHub is the central registry of all connected agents
type AgentHub struct {
	Agents map[string]*ConnectedAgent

	// Register requests from the agents.
	Register chan *ConnectedAgent

	// Unregister requests from agents.
	Unregister chan *ConnectedAgent

	config *config.Config
}

func NewAgentHub(config *config.Config) *AgentHub {
	return &AgentHub{
		Register:   make(chan *ConnectedAgent),
		Unregister: make(chan *ConnectedAgent),
		Agents:     make(map[string]*ConnectedAgent),
		config:     config,
	}
}

func (h *AgentHub) Run() {
	for {
		select {
		case agent := <-h.Register:
			h.Agents[agent.Name] = agent
		case agent := <-h.Unregister:
			if _, ok := h.Agents[agent.Name]; ok {
				delete(h.Agents, agent.Name)
			}
		}
	}
}

func (h *AgentHub) ForceStateSend() {
	for _, a := range h.Agents {
		a.EventChannel <- []byte("{\"action\": \"refetch\"}")
	}
}

func (h *AgentHub) StreamPodLogsSend(namespace string, serviceName string) {
	podlogsRequest := map[string]interface{}{
		"action":      "podLogs",
		"namespace":   namespace,
		"serviceName": serviceName,
	}

	podlogsRequestString, err := json.Marshal(podlogsRequest)
	if err != nil {
		panic(err)
	}

	for _, a := range h.Agents {
		a.EventChannel <- []byte(podlogsRequestString)
	}
}

func (h *AgentHub) StopPodLogs(namespace string, serviceName string) {
	podlogsRequest := map[string]interface{}{
		"action":      "stopPodLogs",
		"namespace":   namespace,
		"serviceName": serviceName,
	}

	podlogsRequestString, err := json.Marshal(podlogsRequest)
	if err != nil {
		panic(err)
	}

	for _, a := range h.Agents {
		a.EventChannel <- []byte(podlogsRequestString)
	}
}

func (a *ConnectedAgent) RepoStacks(repo string) []*api.Stack {
	stacks := []*api.Stack{}

	for _, s := range a.Stacks {
		if repo == s.Repo {
			stacks = append(stacks, s)
		}
	}

	return stacks
}
