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

type AlertStateManager struct {
	notifManager notifications.Manager
	store        store.Store
	waitTime     time.Duration
	thresholds   map[string]threshold
}

func NewAlertStateManager(
	notifManager notifications.Manager,
	store store.Store,
	alertEvaluationFrequencySeconds int,
	thresholds map[string]threshold,
) *AlertStateManager {
	return &AlertStateManager{
		notifManager: notifManager,
		store:        store,
		waitTime:     time.Duration(alertEvaluationFrequencySeconds) * time.Second,
		thresholds:   thresholds,
	}
}

func (a AlertStateManager) Run() {
	for {
		a.evaluatePendingAlerts()
		time.Sleep(a.waitTime)
	}
}

func (a AlertStateManager) evaluatePendingAlerts() {
	alerts, err := a.store.AlertsByState(model.PENDING)
	if err != nil {
		logrus.Errorf("couldn't get pending alerts: %s", err)
	}
	for _, alert := range alerts {
		t := thresholdByType(a.thresholds, alert.Type)
		if t != nil && t.Reached(nil, alert) {
			a.notifManager.Broadcast(&notifications.AlertMessage{
				Alert: *alert,
			})

			err := a.store.UpdateAlertState(alert.ID, model.FIRING)
			if err != nil {
				logrus.Errorf("couldn't set firing state for alerts: %s", err)
			}
		}
	}
}

func (a AlertStateManager) TrackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
		currentTime := time.Now().Unix()

		if a.statusNotChanged(podName, pod.Status) {
			continue
		}

		dbPod := &model.Pod{
			Name:       podName,
			Status:     pod.Status,
			StatusDesc: pod.StatusDescription,
		}
		err := a.store.SaveOrUpdatePod(dbPod)
		if err != nil {
			return err
		}

		alerts, err := a.store.RelatedAlerts(podName)
		if err != nil && err != sql.ErrNoRows {
			logrus.Errorf("couldn't get alert from db: %s", err)
			return err
		}
		nonResolvedAlerts := []*model.Alert{}
		for _, a := range alerts {
			if a.Status != model.RESOLVED {
				nonResolvedAlerts = append(nonResolvedAlerts, a)
			}
		}

		if len(nonResolvedAlerts) == 0 {
			if t, ok := a.thresholds[pod.Status]; ok {
				_, err := a.store.CreateAlert(&model.Alert{
					Name:            podName,
					Type:            thresholdType(t),
					DeploymentName:  deploymentName,
					Status:          model.PENDING,
					LastStateChange: currentTime,
				})
				if err != nil {
					return err
				}
			}
		} else {
			for _, nonResolvedAlert := range nonResolvedAlerts {
				t := thresholdByType(a.thresholds, nonResolvedAlert.Type)
				if t.Resolved(dbPod) {
					a.notifManager.Broadcast(&notifications.AlertMessage{
						Alert: *nonResolvedAlert,
					})

					err := a.store.UpdateAlertState(nonResolvedAlert.ID, model.RESOLVED)
					if err != nil {
						logrus.Errorf("couldn't set resolved state for alerts: %s", err)
					}
				}
			}
		}

	}
	return nil
}

func (a AlertStateManager) TrackEvents(events []api.Event) error {
	// for _, event := range events {
	// 	eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
	// 	deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
	// 	alertState := "Pending"
	// 	alertType := "event"

	// 	if a.alreadyAlerted(eventName, alertType) {
	// 		continue
	// 	}

	// 	err := a.store.SaveOrUpdateKubeEvent(&model.KubeEvent{
	// 		Name:       eventName,
	// 		Status:     event.Status,
	// 		StatusDesc: event.StatusDesc,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = a.store.SaveOrUpdateAlert(&model.Alert{
	// 		Type:            alertType,
	// 		Name:            eventName,
	// 		DeploymentName:  deploymentName,
	// 		Status:          alertState,
	// 		StatusDesc:      event.StatusDesc,
	// 		LastStateChange: event.FirstTimestamp,
	// 		Count:           event.Count,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
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
