package streaming

import (
	"encoding/json"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/sirupsen/logrus"
)

// ConnectedAgent represents a connected k8s cluster
type ConnectedAgent struct {
	Name         string           `json:"name"`
	Namespace    string           `json:"namespace"`
	Certificate  []byte           `json:"-"`
	EventChannel chan []byte      `json:"-"`
	Stacks       []*api.Stack     `json:"-"`
	FluxState    *api.FluxState   `json:"-"`
	FluxStatev2  *api.FluxStatev2 `json:"-"`
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

func (h *AgentHub) TriggerImageBuild(imageBuildId string, imageBuildRequest *dx.ImageBuildRequest) {
	for _, a := range h.Agents {
		if a.Name != imageBuildRequest.Env {
			continue
		}

		imageBuildTriggerString, err := json.Marshal(map[string]interface{}{
			"action":  "imageBuildTrigger",
			"buildId": imageBuildId,
			"request": imageBuildRequest,
		})
		if err != nil {
			logrus.Errorf("could not serialize request: %s", err)
			return
		}

		a.EventChannel <- []byte(imageBuildTriggerString)
	}
}

func (h *AgentHub) StreamPodLogsSend(namespace string, deployment string) {
	podlogsRequest := map[string]interface{}{
		"action":         "podLogs",
		"namespace":      namespace,
		"deploymentName": deployment,
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

func (h *AgentHub) DeploymentDetails(namespace string, deployment string) {
	deploymentDetailsRequest := map[string]interface{}{
		"action":         "deploymentDetails",
		"namespace":      namespace,
		"deploymentName": deployment,
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

func (h *AgentHub) PodDetails(namespace string, podName string) {
	deploymentDetailsRequest := map[string]interface{}{
		"action":    "podDetails",
		"namespace": namespace,
		"podName":   podName,
	}

	podDetailsRequestString, err := json.Marshal(deploymentDetailsRequest)
	if err != nil {
		logrus.Errorf("could not serialize request: %s", err)
		return
	}

	for _, a := range h.Agents {
		a.EventChannel <- []byte(podDetailsRequestString)
	}
}

func (h *AgentHub) StopPodLogs(namespace string, deployment string) {
	podlogsRequest := map[string]interface{}{
		"action":         "stopPodLogs",
		"namespace":      namespace,
		"deploymentName": deployment,
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
