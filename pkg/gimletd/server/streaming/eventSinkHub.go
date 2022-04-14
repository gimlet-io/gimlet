package streaming

import (
	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
)

// EventSink represents a connected event sinkk8s cluster
type EventSink struct {
	EventChannel chan interface{} `json:"-"`
}

// AgentHub is the central registry of all connected agents
type EventSinkHub struct {
	EventSinks map[*EventSink]bool

	// Register requests from the sinks
	Register chan *EventSink

	// Unregister requests from sinks
	Unregister chan *EventSink

	config *config.Config
}

func NewEventSinkHub(config *config.Config) *EventSinkHub {
	return &EventSinkHub{
		Register:   make(chan *EventSink),
		Unregister: make(chan *EventSink),
		EventSinks: make(map[*EventSink]bool),
		config:     config,
	}
}

func (h *EventSinkHub) Run() {
	for {
		select {
		case sink := <-h.Register:
			h.EventSinks[sink] = true
		case sink := <-h.Unregister:
			delete(h.EventSinks, sink)
		}
	}
}

func (h *EventSinkHub) BoradcastEvent(gitopsCommit *model.GitopsCommit) {
	for sink := range h.EventSinks {
		sink.EventChannel <- gitopsCommit
	}
}
