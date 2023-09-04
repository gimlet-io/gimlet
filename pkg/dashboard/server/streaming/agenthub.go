package streaming

import (
	"encoding/json"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/sirupsen/logrus"
)

// ConnectedAgent represents a connected k8s cluster
type ConnectedAgent struct {
	Name         string         `json:"name"`
	Namespace    string         `json:"namespace"`
	EventChannel chan []byte    `json:"-"`
	Stacks       []*api.Stack   `json:"-"`
	FluxState    *api.FluxState `json:"-"`
}

type ImageBuildTrigger struct {
	Action        string                `json:"action"`
	DeployRequest dx.MagicDeployRequest `json:"deployRequest"`
	ImageBuildId  string                `json:"imageBuildId"`
	Image         string                `json:"image"`
	Tag           string                `json:"tag"`
	SourcePath    string                `json:"sourcePath"`
}

// AgentHub is the central registry of all connected agents
type AgentHub struct {
	Agents map[string]*ConnectedAgent

	// Register requests from the agents.
	Register chan *ConnectedAgent

	// Unregister requests from agents.
	Unregister chan *ConnectedAgent
}

func NewAgentHub() *AgentHub {
	return &AgentHub{
		Register:   make(chan *ConnectedAgent),
		Unregister: make(chan *ConnectedAgent),
		Agents:     make(map[string]*ConnectedAgent),
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

func (h *AgentHub) TriggerImageBuild(trigger ImageBuildTrigger) {
	for _, a := range h.Agents {
		if a.Name != trigger.DeployRequest.Env {
			continue
		}

		trigger.Action = "imageBuildTrigger"
		imageBuildTriggerString, err := json.Marshal(trigger)
		if err != nil {
			logrus.Errorf("could not serialize request: %s", err)
			return
		}

		a.EventChannel <- []byte(imageBuildTriggerString)
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
		logrus.Errorf("could not serialize request: %s", err)
		return
	}

	for _, a := range h.Agents {
		a.EventChannel <- []byte(podlogsRequestString)
	}
}

func (h *AgentHub) DeploymentDetails(namespace string, serviceName string) {
	deploymentDetailsRequest := map[string]interface{}{
		"action":      "deploymentDetails",
		"namespace":   namespace,
		"serviceName": serviceName,
	}

	deploymentDetailsRequestString, err := json.Marshal(deploymentDetailsRequest)
	if err != nil {
		logrus.Errorf("could not serialize request: %s", err)
		return
	}

	for _, a := range h.Agents {
		a.EventChannel <- []byte(deploymentDetailsRequestString)
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
		logrus.Errorf("could not serialize request: %s", err)
		return
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
