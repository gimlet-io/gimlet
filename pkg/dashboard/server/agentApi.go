package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/agent"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func register(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Streaming not supported"))
		return
	}

	io.WriteString(w, ": ping\n\n")
	flusher.Flush()

	log.Debugf("agent connected: %s/%s", name, namespace)

	eventChannel := make(chan []byte, 10)
	defer func() {
		<-r.Context().Done()
		close(eventChannel)
		log.Debugf("agent disconnected: %s/%s", name, namespace)
	}()

	a := &streaming.ConnectedAgent{Name: name, Namespace: namespace, EventChannel: eventChannel}

	hub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	hub.Register <- a

	clientHub, _ := r.Context().Value("clientHub").(*streaming.ClientHub)
	broadcastAgentConnectedEvent(clientHub, a)

	for {
		select {
		case <-r.Context().Done():
			hub.Unregister <- a
			broadcastAgentDisconnectedEvent(clientHub, a)
			return
		case <-time.After(time.Second * 30):
			io.WriteString(w, ": ping\n\n")
			flusher.Flush()
		case buf, ok := <-eventChannel:
			if ok {
				io.WriteString(w, "data: ")
				w.Write(buf)
				io.WriteString(w, "\n\n")
				flusher.Flush()
			}
		}
	}
}

func broadcastAgentConnectedEvent(clientHub *streaming.ClientHub, a *streaming.ConnectedAgent) {
	jsonString, _ := json.Marshal(streaming.AgentConnectedEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.AgentConnectedEventString},
		Agent:          *a,
	})
	clientHub.Broadcast <- jsonString
}

func broadcastAgentDisconnectedEvent(clientHub *streaming.ClientHub, a *streaming.ConnectedAgent) {
	jsonString, _ := json.Marshal(streaming.AgentDisconnectedEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.AgentDisconnectedEventString},
		Agent:          *a,
	})
	clientHub.Broadcast <- jsonString
}

func state(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	var stacks []api.Stack
	err := json.NewDecoder(r.Body).Decode(&stacks)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
	agent := agentHub.Agents[name]
	if agent == nil {
		time.Sleep(1 * time.Second) // Agenthub has a race condition. Registration is not done when the client sends the state
		agent = agentHub.Agents[name]
	}

	stackPointers := []*api.Stack{}
	for _, s := range stacks {
		copy := s       // needed as the address of s is constant in the for loop
		copy.Env = name // making the service aware of its env
		stackPointers = append(stackPointers, &copy)
	}
	agent.Stacks = stackPointers

	envs := []*api.Env{{
		Name:   name,
		Stacks: stackPointers,
	}}

	err = decorateDeployments(r.Context(), envs)
	if err != nil {
		logrus.Errorf("cannot decorate deployments: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	clientHub, _ := r.Context().Value("clientHub").(*streaming.ClientHub)
	jsonString, _ := json.Marshal(streaming.EnvsUpdatedEvent{
		StreamingEvent: streaming.StreamingEvent{Event: streaming.EnvsUpdatedEventString},
		Envs:           envs,
	})
	clientHub.Broadcast <- jsonString
}

func update(w http.ResponseWriter, r *http.Request) {
	var update api.StackUpdate
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

	update = decorateDeploymentUpdateWithCommitMessage(update, r)

	poorMansNewServiceHandler(update, r)

	clientHub, _ := r.Context().Value("clientHub").(*streaming.ClientHub)
	jsonString, _ := json.Marshal(update)
	clientHub.Broadcast <- jsonString
}

func poorMansNewServiceHandler(update api.StackUpdate, r *http.Request) {
	// delete it when properly handling svc created event in agents,
	// and covered all eventual consistency cases
	if update.Event == agent.EventDeploymentCreated {
		agentHub, _ := r.Context().Value("agentHub").(*streaming.AgentHub)
		go func() {
			time.Sleep(100 * time.Millisecond)
			agentHub.ForceStateSend()
		}()
	}
}

func decorateDeploymentUpdateWithCommitMessage(update api.StackUpdate, r *http.Request) api.StackUpdate {
	if update.Event == agent.EventDeploymentUpdated ||
		update.Event == agent.EventDeploymentCreated {
		dao := r.Context().Value("store").(*store.Store)

		dbCommits, err := dao.CommitsByRepoAndSHA(update.Repo, []string{update.SHA})
		if err != nil {
			logrus.Warnf("cannot get commits from db %s", err)
		}
		if len(dbCommits) == 1 {
			update.CommitMessage = dbCommits[0].Message
		}
	}

	return update
}
