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

func (a alertStateManager) TrackPods(pods []*api.Pod) error {
	return a.trackPods(pods)
}

func (a alertStateManager) TrackEvents(events []api.Event) error {
	return a.trackEvents(events)
}

func (a alertStateManager) Run() {
	for {
		a.agentHub.GetEvents()

		thresholdPods, err := a.podThresholdsFromAlerts("pod")
		if err != nil {
			logrus.Errorf("could't save or update pod: %s", err)
		}

		err = a.setFiringState(thresholdPods)
		if err != nil {
			logrus.Errorf("could't save or update pod: %s", err)
		}

		thresholdEvents, err := a.eventThresholdsFromAlerts("event")
		if err != nil {
			logrus.Errorf("could't save or update pod: %s", err)
		}

		err = a.setFiringState(thresholdEvents)
		if err != nil {
			logrus.Errorf("could't save or update pod: %s", err)
		}

		time.Sleep(a.waitTime * time.Minute)
	}
}

func (a alertStateManager) DeletePod(podName string) error {
	return a.store.DeletePod(podName)
}

func (a alertStateManager) DeleteEvent(name string) error {
	return a.store.DeleteEvent(name)
}

func (a alertStateManager) Alerts() ([]*model.Alert, error) {
	return a.store.Alerts()
}

func (a alertStateManager) trackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
		currentTime := time.Now().Unix()
		alertState := "Pending"
		alertType := "pod"

		if a.alreadyAlerted(podName, alertType) || a.statusNotChanged(podName, pod.Status) {
			continue
		}

		if !podErrorState(pod.Status) {
			alertState = "OK"
		}

		err := a.store.SaveOrUpdatePod(&model.Pod{
			Name:       podName,
			Status:     pod.Status,
			StatusDesc: pod.StatusDescription,
		})
		if err != nil {
			return err
		}

		err = a.store.SaveAlert(&model.Alert{
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

func (a alertStateManager) trackEvents(events []api.Event) error {
	for _, event := range events {
		eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
		alertState := "Pending"
		alertType := "event"

		if a.alreadyAlerted(eventName, alertType) {
			continue
		}

		err := a.store.SaveOrUpdateEvent(&model.Event{
			Name:       eventName,
			Status:     event.Status,
			StatusDesc: event.StatusDesc,
		})
		if err != nil {
			return err
		}

		err = a.store.SaveAlert(&model.Alert{
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

func (a alertStateManager) podThresholdsFromAlerts(alertType string) ([]threshold, error) {
	var thresholds []threshold
	alerts, err := a.store.PendingAlertsByType(alertType)
	if err != nil {
		return thresholds, err
	}

	for _, alert := range alerts {
		thresholds = append(thresholds, thresholdFromPod(*alert, a.waitTime))
	}
	return thresholds, nil
}

func (a alertStateManager) eventThresholdsFromAlerts(alertType string) ([]threshold, error) {
	var thresholds []threshold
	alerts, err := a.store.PendingAlertsByType(alertType)
	if err != nil {
		return thresholds, err
	}

	for _, alert := range alerts {
		thresholds = append(thresholds, thresholdFromEvent(*alert))
	}
	return thresholds, nil
}

func (a alertStateManager) setFiringState(thresholds []threshold) error {
	for _, t := range thresholds {
		if t.isFired() {
			msg := notifications.MessageFromAlert(t.model())
			a.notifManager.Broadcast(msg)

			err := a.store.SaveAlert(&model.Alert{
				// TODO update alert with status "Firing"
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a alertStateManager) alreadyAlerted(name string, alertType string) bool {
	alert, err := a.store.Alert(name, alertType)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return alert.Status == "Fired"
}

func (a alertStateManager) statusNotChanged(podName string, podStatus string) bool {
	podFromDb, err := a.store.Pod(podName)
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
