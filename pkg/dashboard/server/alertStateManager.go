package server

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

type alertStateManager struct {
	notifManager notifications.Manager
	agentHub     *streaming.AgentHub
	store        store.Store
	waitTime     time.Duration
}

func NewAlertStateManager(notifManager notifications.Manager, agentHub *streaming.AgentHub, store store.Store, waitTime time.Duration) *alertStateManager {
	return &alertStateManager{notifManager: notifManager, agentHub: agentHub, store: store, waitTime: waitTime}
}

func (p alertStateManager) TrackPods(pods []*api.Pod) error {
	return p.trackPods(pods)
}

func (p alertStateManager) TrackEvents(events []api.Event) error {
	return p.trackEvents(events)
}

func (p alertStateManager) Run() {
	for {
		p.agentHub.GetEvents()
		// TODO error handling
		p.setFiringStateForPods()
		p.setFiringStateForEvents()
		time.Sleep(p.waitTime * time.Minute)
	}
}

func (p alertStateManager) DeletePod(podName string) error {
	return p.store.DeletePod(podName)
}

func (p alertStateManager) DeleteEvent(name string) error {
	return p.store.DeleteEvent(name)
}

func (p alertStateManager) Alerts() ([]*model.Alert, error) {
	return p.store.Alerts()
}

func (p alertStateManager) trackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
		currentTime := time.Now().Unix()
		alertState := "Pending"
		alertType := "pod"

		if p.alreadyAlerted(podName, alertType) || p.statusNotChanged(podName, pod.Status) {
			continue
		}

		if !podErrorState(pod.Status) {
			alertState = "OK"
		}

		err := p.store.SaveOrUpdatePod(&model.Pod{
			Name:       podName,
			Status:     pod.Status,
			StatusDesc: pod.StatusDescription,
		})
		if err != nil {
			return err
		}

		err = p.store.SaveAlert(&model.Alert{
			Type:           alertType,
			Name:           podName,
			DeploymentName: deploymentName,
			Status:         alertState,
			StatusDesc:     pod.StatusDescription,
			Fired:          currentTime,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p alertStateManager) trackEvents(events []api.Event) error {
	for _, event := range events {
		eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
		alertState := "Pending"
		alertType := "event"

		if p.alreadyAlerted(eventName, alertType) {
			continue
		}

		err := p.store.SaveOrUpdateEvent(&model.Event{
			Name:       eventName,
			Status:     event.Status,
			StatusDesc: event.StatusDesc,
		})
		if err != nil {
			return err
		}

		err = p.store.SaveAlert(&model.Alert{
			Type:           alertType,
			Name:           eventName,
			DeploymentName: deploymentName,
			Status:         alertState,
			StatusDesc:     event.StatusDesc,
			Fired:          event.FirstTimestamp,
			Count:          event.Count,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p alertStateManager) setFiringStateForPods() error {
	pods, err := p.store.PendingAlertsByType("pod")
	if err != nil {
		return err
	}

	for _, pod := range pods {
		t := p.thresholdFromPod(*pod)

		if t.isFired() {
			msg := notifications.MessageFromAlert(*pod)
			p.notifManager.Broadcast(msg)

			err := p.store.SaveAlert(&model.Alert{
				// TODO update alert with status "Firing"
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p alertStateManager) setFiringStateForEvents() error {
	events, err := p.store.PendingAlertsByType("event")
	if err != nil {
		return err
	}

	for _, event := range events {
		t := p.thresholdFromEvent(*event)

		if t.isFired() {
			msg := notifications.MessageFromAlert(*event)
			p.notifManager.Broadcast(msg)

			err := p.store.SaveAlert(&model.Alert{
				// TODO update alert with status "Firing"
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p alertStateManager) thresholdFromPod(pod model.Alert) threshold {
	return &podStrategy{
		timestamp: pod.Fired,
		waitTime:  p.waitTime,
	}
}

func (p alertStateManager) thresholdFromEvent(event model.Alert) threshold {
	return &eventStrategy{
		timestamp:              event.Fired,
		count:                  event.Count,
		expectedCountPerMinute: 1, // TODO make it configurable
		expectedCount:          6, // TODO make it configurable
	}
}

func (p alertStateManager) alreadyAlerted(name string, alertType string) bool {
	alert, err := p.store.Alert(name, alertType)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return alert.Status == "Fired"
}

func (p alertStateManager) statusNotChanged(podName string, podStatus string) bool {
	podFromDb, err := p.store.Pod(podName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return podStatus == podFromDb.Status
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
