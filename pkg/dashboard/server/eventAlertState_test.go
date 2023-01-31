package server

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
)

func TestTrackEvents(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	event1 := api.Event{Namespace: "ns1", Name: "pod1"}
	event2 := api.Event{Namespace: "ns1", Name: "pod2"}
	events := []api.Event{event1, event2}

	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
	p.trackEvents(events)

	expected := []model.Event{
		{Name: "ns1/pod1", AlertState: "Pending"},
		{Name: "ns1/pod2", AlertState: "Pending"},
	}
	for _, event := range expected {
		e, _ := store.Event(event.Name)

		assert.Equal(t, event.Name, e.Name)
		assert.Equal(t, event.AlertState, e.AlertState)
	}
}

func TestSetFiringStatesForEvents(t *testing.T) {
	store := store.NewTest()
	defer func() {
		store.Close()
	}()

	dummyNotificationsManager := notifications.NewDummyManager()
	event1 := api.Event{Namespace: "ns1", Name: "pod1", FirstTimestamp: 0, Count: 0}
	event2 := api.Event{Namespace: "ns1", Name: "pod2", FirstTimestamp: 0, Count: 0}
	events := []api.Event{event1, event2}

	p := NewAlertStateManager(dummyNotificationsManager, *store, 2)
	p.trackEvents(events)
	p.setFiringStateForEvents()

	expected := []model.Event{
		{Name: "ns1/pod1", AlertState: "Pending"},
		{Name: "ns1/pod2", AlertState: "Firing"},
	}
	for _, event := range expected {
		e, _ := store.Event(event.Name)

		assert.Equal(t, event.Name, e.Name)
		assert.Equal(t, event.AlertState, e.AlertState)
	}
}
