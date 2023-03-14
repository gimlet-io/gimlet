package alert

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

func getExpectedNumbers() map[string]expected {
	return map[string]expected{
		"ImagePullBackOff": {
			waitTime:       2,
			Count:          6,
			CountPerMinute: 1,
		},
		"Failed": {
			waitTime:       2,
			Count:          6,
			CountPerMinute: 1,
		},
		// TODO insert different type of errors
	}
}

type expected struct {
	waitTime       time.Duration
	Count          int32
	CountPerMinute float64
}

type AlertStateManager struct {
	notifManager notifications.Manager
	store        store.Store
	waitTime     time.Duration
}

func NewAlertStateManager(notifManager notifications.Manager, store store.Store, waitTime time.Duration) *AlertStateManager {
	return &AlertStateManager{notifManager: notifManager, store: store, waitTime: waitTime}
}

func (a AlertStateManager) Run() {
	for {
		var thresholds []threshold
		alerts, err := a.store.PendingAlerts()
		if err != nil {
			logrus.Errorf("couldn't get pending alerts: %s", err)
		}
		for _, alert := range alerts {
			status, err := a.status(alert.Name, alert.Type)
			if err != nil {
				logrus.Errorf("couldn't get status from alert: %s", err)
				continue
			}
			expected := getExpectedNumbers()[status]
			thresholds = append(thresholds, ToThreshold(alert, expected.waitTime, expected.Count, expected.CountPerMinute))
		}

		err = a.setFiringState(thresholds)
		if err != nil {
			logrus.Errorf("couldn't set firing state for alerts: %s", err)
		}

		time.Sleep(a.waitTime * time.Minute)
	}
}

func (a AlertStateManager) TrackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
		currentTime := time.Now().Unix()
		alertState := "Pending"
		alertType := "pod"

		if a.alreadyAlerted(podName, alertType) || a.statusNotChanged(podName, pod.Status) {
			continue
		}

		err := a.store.SaveOrUpdatePod(&model.Pod{
			Name:       podName,
			Status:     pod.Status,
			StatusDesc: pod.StatusDescription,
		})
		if err != nil {
			return err
		}

		if podErrorState(pod.Status) {
			err := a.store.SaveOrUpdateAlert(&model.Alert{
				Type:            alertType,
				Name:            podName,
				DeploymentName:  deploymentName,
				Status:          alertState,
				StatusDesc:      pod.StatusDescription,
				LastStateChange: currentTime,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a AlertStateManager) TrackEvents(events []api.Event) error {
	for _, event := range events {
		eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
		alertState := "Pending"
		alertType := "event"

		if a.alreadyAlerted(eventName, alertType) {
			continue
		}

		err := a.store.SaveOrUpdateKubeEvent(&model.KubeEvent{
			Name:       eventName,
			Status:     event.Status,
			StatusDesc: event.StatusDesc,
		})
		if err != nil {
			return err
		}

		err = a.store.SaveOrUpdateAlert(&model.Alert{
			Type:            alertType,
			Name:            eventName,
			DeploymentName:  deploymentName,
			Status:          alertState,
			StatusDesc:      event.StatusDesc,
			LastStateChange: event.FirstTimestamp,
			Count:           event.Count,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a AlertStateManager) setFiringState(thresholds []threshold) error {
	for _, t := range thresholds {
		if t.isFired() {
			alert := t.toAlert()
			msg := notifications.MessageFromAlert(alert)
			a.notifManager.Broadcast(msg)

			err := a.store.SaveOrUpdateAlert(&model.Alert{
				Type:            alert.Type,
				Name:            alert.Name,
				DeploymentName:  alert.DeploymentName,
				Status:          "Firing",
				StatusDesc:      alert.StatusDesc,
				LastStateChange: time.Now().Unix(),
				Count:           alert.Count,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a AlertStateManager) status(name string, alertType string) (string, error) {
	if alertType == "pod" {
		pod, err := a.store.Pod(name)
		if err != nil {
			return "", err
		}
		return pod.Status, nil
	}

	event, err := a.store.KubeEvent(name)
	if err != nil {
		return "", err
	}
	return event.Status, nil
}

func (a AlertStateManager) alreadyAlerted(name string, alertType string) bool {
	alert, err := a.store.Alert(name, alertType)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("couldn't get pod from db: %s", err)
		return false
	}

	return alert.Status == "Firing"
}

func (a AlertStateManager) statusNotChanged(podName string, podStatus string) bool {
	podFromDb, err := a.store.Pod(podName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("couldn't get pod from db: %s", err)
		return false
	}

	return podStatus == podFromDb.Status
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
