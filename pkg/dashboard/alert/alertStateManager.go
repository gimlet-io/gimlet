package alert

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

type AlertStateManager struct {
	notifManager notifications.Manager
	clientHub    *streaming.ClientHub
	store        store.Store
	waitTime     time.Duration
	thresholds   map[string]threshold
}

func NewAlertStateManager(
	notifManager notifications.Manager,
	clientHub *streaming.ClientHub,
	store store.Store,
	alertEvaluationFrequencySeconds int,
	thresholds map[string]threshold,
) *AlertStateManager {
	return &AlertStateManager{
		notifManager: notifManager,
		clientHub:    clientHub,
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

			a.broadcast(&model.Alert{
				Type:           alert.Type,
				ObjectName:     alert.ObjectName,
				DeploymentName: alert.DeploymentName,
				Status:         model.FIRING,
				PendingAt:      alert.PendingAt,
				FiredAt:        time.Now().Unix(),
			},
				streaming.AlertFiredEventString,
			)
		}
	}
}

// TrackDeploymentPods tracks all pods state and related alerts for a deployment
// This is authoritive tracking, so call it with all the pods related to a deployment
func (a AlertStateManager) TrackDeploymentPods(pods []*api.Pod) error {
	for _, pod := range pods {
		err := a.TrackPod(pod)
		if err != nil {
			return err
		}
	}

	if len(pods) == 0 {
		return nil
	}

	deploymentName := fmt.Sprintf("%s/%s", pods[0].Namespace, pods[0].DeploymentName)
	alerts, err := a.store.AlertsByDeployment(deploymentName)
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("couldn't get alert from db: %s", err)
		return err
	}

	// resolve alerts related to non-existing pods
	for _, alert := range alerts {
		if alert.Status == model.RESOLVED {
			continue
		}
		if podExists(pods, alert.ObjectName) {
			continue
		}
		err := a.DeletePod(alert.ObjectName)
		if err != nil {
			return err
		}

	}

	return nil
}

func podExists(pods []*api.Pod, pod string) bool {
	for _, p := range pods {
		if fmt.Sprintf("%s/%s", p.Namespace, p.Name) == pod {
			return true
		}
	}

	return false
}

// TrackPod tracks a pod state and related alerts
func (a AlertStateManager) TrackPod(pod *api.Pod) error {
	podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	deploymentName := fmt.Sprintf("%s/%s", pod.Namespace, pod.DeploymentName)
	currentTime := time.Now().Unix()

	if a.statusNotChanged(podName, pod.Status) {
		return nil
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

	if t, ok := a.thresholds[pod.Status]; ok {
		alertToCreate := &model.Alert{
			ObjectName:     podName,
			Type:           thresholdType(t),
			DeploymentName: deploymentName,
			Status:         model.PENDING,
			PendingAt:      currentTime,
		}
		if !alertExists(nonResolvedAlerts, alertToCreate) {
			_, err := a.store.CreateAlert(alertToCreate)
			if err != nil {
				return err
			}

			a.broadcast(alertToCreate, streaming.AlertPendingEventString)
		}
	}

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

			a.broadcast(&model.Alert{
				Type:           nonResolvedAlert.Type,
				ObjectName:     nonResolvedAlert.ObjectName,
				DeploymentName: nonResolvedAlert.DeploymentName,
				Status:         model.RESOLVED,
				PendingAt:      nonResolvedAlert.PendingAt,
				FiredAt:        nonResolvedAlert.FiredAt,
				ResolvedAt:     time.Now().Unix(),
			},
				streaming.AlertResolvedEventString,
			)
		}
	}

	return nil
}

func (a AlertStateManager) DeletePod(podName string) error {
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

	for _, nonResolvedAlert := range nonResolvedAlerts {
		err := a.store.UpdateAlertState(nonResolvedAlert.ID, model.RESOLVED)
		if err != nil {
			logrus.Errorf("couldn't set resolved state for alerts: %s", err)
		}

		a.broadcast(&model.Alert{
			Type:           nonResolvedAlert.Type,
			ObjectName:     nonResolvedAlert.ObjectName,
			DeploymentName: nonResolvedAlert.DeploymentName,
			Status:         model.RESOLVED,
			PendingAt:      nonResolvedAlert.PendingAt,
			FiredAt:        nonResolvedAlert.FiredAt,
			ResolvedAt:     time.Now().Unix(),
		},
			streaming.AlertResolvedEventString,
		)
	}

	return a.store.DeletePod(podName)
}

func alertExists(nonResolvedAlerts []*model.Alert, alert *model.Alert) bool {
	for _, nonResolvedAlert := range nonResolvedAlerts {
		if nonResolvedAlert.ObjectName == alert.ObjectName && nonResolvedAlert.Type == alert.Type {
			return true
		}
	}

	return false
}

func (a AlertStateManager) broadcast(alert *model.Alert, event string) {
	if a.clientHub == nil {
		return
	}
	jsonString, _ := json.Marshal(streaming.AlertEvent{
		Alert:          alert,
		StreamingEvent: streaming.StreamingEvent{Event: event},
	})
	a.clientHub.Broadcast <- jsonString
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
