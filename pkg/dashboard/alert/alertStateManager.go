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

var podTresholds map[string]podThreshold = map[string]podThreshold{
	"ImagePullBackOff": {
		waitTime: 2,
	},
}

var eventTresholds map[string]eventThreshold = map[string]eventThreshold{
	"Failed": {
		expectedCountPerMinute: 1,
		expectedCount:          6,
	},
}

type AlertStateManager struct {
	notifManager notifications.Manager
	store        store.Store
	waitTime     time.Duration
}

func NewAlertStateManager(
	notifManager notifications.Manager,
	store store.Store,
	alertEvaluationFrequencySeconds int,
) *AlertStateManager {
	return &AlertStateManager{
		notifManager: notifManager,
		store:        store,
		waitTime:     time.Duration(alertEvaluationFrequencySeconds) * time.Second,
	}
}

func (a AlertStateManager) Run() {
	for {
		a.evaluatePendingAlerts()
		time.Sleep(a.waitTime)
	}
}

func (a AlertStateManager) evaluatePendingAlerts() {
	alerts, err := a.store.PendingAlerts()
	if err != nil {
		logrus.Errorf("couldn't get pending alerts: %s", err)
	}

	for _, alert := range alerts {
		var t threshold
		var relatedObject interface{}
		switch alert.Type {
		case model.ALERT_OBJECT_TYPE_POD:
			pod, err := a.store.Pod(alert.Name)
			if err != nil {
				logrus.Errorf("could not get related object for alert: %s", err)
				continue
			}
			relatedObject = pod
			if val, ok := podTresholds[pod.Status]; ok {
				t = val
			} else {
				t = noopThreshold{}
			}
		case model.ALERT_OBJECT_TYPE_EVENT:
			event, err := a.store.KubeEvent(alert.Name)
			if err != nil {
				logrus.Errorf("could not get related object for alert: %s", err)
				continue
			}
			relatedObject = event
			if val, ok := eventTresholds[event.Status]; ok {
				t = val
			} else {
				t = noopThreshold{}
			}
		}

		if t.isReached(relatedObject, alert) {
			alert.Status = model.ALERT_STATE_FIRING
			err := a.store.UpdateAlert(alert)
			if err != nil {
				logrus.Errorf("could not update alert state: %s", err)
				continue
			}
			msg := notifications.MessageFromAlert(*alert)
			a.notifManager.Broadcast(msg)
		}
	}
}

func (a AlertStateManager) TrackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
		relatedAlert, err := a.relatedAlert(podName, model.ALERT_OBJECT_TYPE_POD)
		if err != nil {
			return err
		}

		if a.statusNotChanged(podName, pod.Status) {
			continue
		}

		if relatedAlert != nil && relatedAlert.IsFiring() {
			continue
		}

		storedPod, err := a.store.SaveOrUpdatePod(&model.Pod{
			Name:       podName,
			Status:     pod.Status,
			StatusDesc: pod.StatusDescription,
		})
		if err != nil {
			return err
		}

		if storedPod.IsInErrorState() {
			if relatedAlert != nil {
				relatedAlert.Status = model.ALERT_STATE_PENDING
				err = a.store.UpdateAlert(relatedAlert)
				if err != nil {
					return err
				}
			} else {
				err := a.store.CreateAlert(&model.Alert{
					Type:            model.ALERT_OBJECT_TYPE_POD,
					Name:            podName,
					DeploymentName:  deploymentName,
					Status:          model.ALERT_STATE_PENDING,
					StatusDesc:      pod.StatusDescription,
					LastStateChange: time.Now().Unix(),
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a AlertStateManager) TrackEvents(events []api.Event) error {
	for _, event := range events {
		eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
		relatedAlert, err := a.relatedAlert(eventName, model.ALERT_OBJECT_TYPE_EVENT)
		if err != nil {
			return err
		}

		if a.statusNotChanged(eventName, event.Status) {
			continue
		}

		if relatedAlert != nil && relatedAlert.IsFiring() {
			continue
		}

		err = a.store.SaveOrUpdateKubeEvent(&model.KubeEvent{
			Name:       eventName,
			Status:     event.Status,
			StatusDesc: event.StatusDesc,
			Count:      int(event.Count),
		})
		if err != nil {
			return err
		}

		if relatedAlert != nil {
			relatedAlert.Status = model.ALERT_STATE_PENDING
			err = a.store.UpdateAlert(relatedAlert)
			if err != nil {
				return err
			}
		} else {
			err := a.store.CreateAlert(&model.Alert{
				Type:            model.ALERT_OBJECT_TYPE_EVENT,
				Name:            eventName,
				DeploymentName:  deploymentName,
				Status:          model.ALERT_STATE_PENDING,
				StatusDesc:      event.StatusDesc,
				LastStateChange: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a AlertStateManager) relatedAlert(objectName string, objectType string) (*model.Alert, error) {
	alert, err := a.store.Alert(objectName, objectType)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("couldn't get alert for pod from db: %s", err)
	}
	return alert, nil
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
