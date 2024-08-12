package worker

import (
	"encoding/json"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

type EventsWorker struct {
	dao       *store.Store
	clientHub *streaming.ClientHub
	events    chan *model.Event
}

func NewEventsWorker(
	dao *store.Store,
	clientHub *streaming.ClientHub,
) *EventsWorker {
	return &EventsWorker{
		dao:       dao,
		clientHub: clientHub,
		events:    make(chan *model.Event, 100),
	}
}

func (a *EventsWorker) Run() {
	a.dao.SubscribeToEventCreated(func(event *model.Event) {
		a.events <- event
	})

	a.dao.SubscribeToEventUpdated(func(event *model.Event) {
		a.events <- event
	})

	for {
		event := <-a.events
		commitEvent := server.AsCommitEvent(event)

		eventsString, err := json.Marshal(streaming.CommitEvent{
			StreamingEvent: streaming.StreamingEvent{Event: streaming.CommitEventString},
			CommitEvent:    commitEvent,
		})
		if err != nil {
			logrus.Warnf("cannot serialize commits: %s", err)
			continue
		}
		a.clientHub.Broadcast <- eventsString
	}
}
