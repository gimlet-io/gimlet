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
	host         string
}

func NewAlertStateManager(
	notifManager notifications.Manager,
	clientHub *streaming.ClientHub,
	store store.Store,
	alertEvaluationFrequencySeconds int,
	thresholds map[string]threshold,
	host string,
) *AlertStateManager {
	return &AlertStateManager{
		notifManager: notifManager,
		clientHub:    clientHub,
		store:        store,
		waitTime:     time.Duration(alertEvaluationFrequencySeconds) * time.Second,
		thresholds:   thresholds,
		host:         host,
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
		t := ThresholdByType(a.thresholds, alert.Type)
		if t == nil { // resolving unknown pending alerts (sometimes we deprecate alerts)
			alert.SetResolved()
			err := a.store.UpdateAlertState(alert)
			if err != nil {
				logrus.Errorf("couldn't set resolved state for alerts: %s", err)
			}
		}
		if t != nil && t.Reached(nil, alert) {
			alert.SetFiring()
			err := a.store.UpdateAlertState(alert)
			if err != nil {
				logrus.Errorf("couldn't set firing state for alerts: %s", err)
			}

			silencedUntil, err := a.store.DeploymentSilencedUntil(alert.DeploymentName, alert.Type)
			if err != nil {
				logrus.Errorf("couldn't get deployment silenced until: %s", err)
			}

			apiAlert := api.NewAlert(alert, t.Text(), t.Name(), silencedUntil)
			if a.alertsSilenced(alert.DeploymentName, alert.Type) {
				if a.alertsSilenced(alert.DeploymentName, alert.Type) {
					a.notifManager.Broadcast(&notifications.AlertMessage{
						Alert:         *apiAlert,
						ImChannelId:   alert.ImChannelId,
						DeploymentUrl: alert.DeploymentUrl,
					})
				}
			}
			a.broadcast(apiAlert, streaming.AlertFiredEventString)
		}
	}
}

// TrackDeploymentPods tracks all pods state and related alerts for a deployment
// This is authoritive tracking, so call it with all the pods related to a deployment
func (a AlertStateManager) TrackDeploymentPods(pods []*api.Pod, repoName string, envName string) error {
	for _, pod := range pods {
		err := a.TrackPod(pod, repoName, envName)
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
func (a AlertStateManager) TrackPod(pod *api.Pod, repoName string, envName string) error {
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
			ImChannelId:    pod.ImChannelId,
			DeploymentUrl:  fmt.Sprintf("%s/repo/%s/%s/%s", a.host, repoName, envName, pod.DeploymentName),
		}
		if !alertExists(nonResolvedAlerts, alertToCreate) {
			_, err := a.store.CreateAlert(alertToCreate)
			if err != nil {
				return err
			}
			silencedUntil, err := a.store.DeploymentSilencedUntil(alertToCreate.DeploymentName, alertToCreate.Type)
			if err != nil {
				logrus.Errorf("couldn't get deployment silenced until: %s", err)
			}
			apiAlert := api.NewAlert(alertToCreate, t.Text(), t.Name(), silencedUntil)
			a.broadcast(apiAlert, streaming.AlertPendingEventString)
		}
	}

	for _, nonResolvedAlert := range nonResolvedAlerts {
		t := ThresholdByType(a.thresholds, nonResolvedAlert.Type)
		if t != nil && t.Resolved(dbPod) {
			previousState := nonResolvedAlert.Status
			nonResolvedAlert.SetResolved()
			err := a.store.UpdateAlertState(nonResolvedAlert)
			if err != nil {
				logrus.Errorf("couldn't set resolved state for alerts: %s", err)
			}

			silencedUntil, err := a.store.DeploymentSilencedUntil(nonResolvedAlert.DeploymentName, nonResolvedAlert.Type)
			if err != nil {
				logrus.Errorf("couldn't get deployment silenced until: %s", err)
			}

			apiAlert := api.NewAlert(nonResolvedAlert, t.Text(), t.Name(), silencedUntil)
			if a.alertsSilenced(nonResolvedAlert.DeploymentName, nonResolvedAlert.Type) {
				if previousState == model.FIRING { // don't notify people about pending then resolved alerts
					a.notifManager.Broadcast(&notifications.AlertMessage{
						Alert:       *apiAlert,
						ImChannelId: pod.ImChannelId,
					})
				}
			}
			a.broadcast(apiAlert, streaming.AlertResolvedEventString)
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
		previousState := nonResolvedAlert.Status
		nonResolvedAlert.SetResolved()
		err := a.store.UpdateAlertState(nonResolvedAlert)
		if err != nil {
			logrus.Errorf("couldn't set resolved state for alerts: %s", err)
		}

		apiAlert := api.NewAlert(nonResolvedAlert, "", "", 0)
		if a.alertsSilenced(nonResolvedAlert.DeploymentName, nonResolvedAlert.Type) {
			if previousState == model.FIRING { // don't notify people about pending then resolved alerts
				a.notifManager.Broadcast(&notifications.AlertMessage{
					Alert:       *apiAlert,
					ImChannelId: nonResolvedAlert.ImChannelId,
				})
			}
		}
		a.broadcast(apiAlert, streaming.AlertResolvedEventString)
	}

	return a.store.DeletePod(podName)
}

func alertExists(existingAlerts []*model.Alert, alert *model.Alert) bool {
	for _, existingAlert := range existingAlerts {
		if objectsMatch(existingAlert, alert) {
			if typesMatch(existingAlert, alert) {
				return true
			} else if isCrashLoopBackOffSpecialCase(existingAlert, alert) {
				return true
			}
		}
	}

	return false
}

func objectsMatch(first *model.Alert, second *model.Alert) bool {
	return first.ObjectName == second.ObjectName
}

func typesMatch(first *model.Alert, second *model.Alert) bool {
	return first.Type == second.Type
}

// isCrashLoopBackOffSpecialCase returns true if the examined alert is a crashLoopBackOff
// and an OOMKilledAlert already exists. OOMKilled is a special kind of CrashloopBackOff
func isCrashLoopBackOffSpecialCase(existingAlert, a *model.Alert) bool {
	return a.Type == "crashLoopBackOffThreshold" && existingAlert.Type == "oomKilledThreshold"
}

func (a AlertStateManager) broadcast(alert *api.Alert, event string) {
	if a.clientHub == nil {
		return
	}
	jsonString, _ := json.Marshal(streaming.AlertEvent{
		Alert:          alert,
		StreamingEvent: streaming.StreamingEvent{Event: event},
	})
	a.clientHub.Broadcast <- jsonString
}

func (a AlertStateManager) alertsSilenced(deploymentName string, alertType string) bool {
	object := fmt.Sprintf("%s-%s", deploymentName, alertType)
	storedObject, err := a.store.KeyValue(object)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("cannot get key value")
		return false
	}

	var silencedUntil *time.Time
	t, err := time.Parse(time.RFC3339, storedObject.Value)
	if err != nil {
		logrus.Errorf("cannot parse until date %s", err)
		return false
	}
	silencedUntil = &t

	return time.Now().Before(*silencedUntil)
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
