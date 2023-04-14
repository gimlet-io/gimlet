package alert

import (
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

const ImagePullBackOff = "ImagePullBackOff"
const ErrImagePull = "ErrImagePull"
const CreateContainerConfigError = "CreateContainerConfigError"
const CrashLoopBackOff = "CrashLoopBackOff"

// https://kukulinski.com/10-most-common-reasons-kubernetes-deployments-fail-part-2/
// things like cluster fill aka forever pending
// and container creating forever as disks are not mounting
// events

var thresholds map[string]threshold = map[string]threshold{
	CrashLoopBackOff: podThreshold{
		waitTime: 60,
	},
	ImagePullBackOff: podThreshold{
		waitTime: 60,
	},
	CreateContainerConfigError: zeroThreshold{},
	"Failed": eventThreshold{
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
	alerts, err := a.store.FiringAlerts()
	for _, alert := range alerts {
		fmt.Printf("🔥 %s (%s) - %s\n", alert.ObjectName, alert.ObjectStatus, alert.Status)
	}

	alerts, err = a.store.PendingAlerts()
	if err != nil {
		logrus.Errorf("couldn't get pending alerts: %s", err)
	}

	for _, alert := range alerts {
		fmt.Printf("%s (%s) - %s\n", alert.ObjectName, alert.ObjectStatus, alert.Status)
		var t threshold
		if val, ok := thresholds[alert.ObjectStatus]; ok {
			t = val
		} else {
			err := a.store.DeleteAlert(alert.ID)
			if err != nil {
				logrus.Errorf("could not delete alert: %s", err)
			}
			continue
		}

		if t.isReached(nil, alert) {
			alert.Status = model.ALERT_STATE_FIRING
			err := a.store.UpdateAlertStatus(alert.ID, alert.Status)
			if err != nil {
				logrus.Errorf("could not update alert state: %s", err)
				continue
			}
			msg := notifications.MessageFromAlert(*alert)
			a.notifManager.Broadcast(msg)
		}
	}
}

func (a AlertStateManager) PodDeleted(podName string) error {
	relatedAlerts, err := a.relatedAlerts(podName, model.ALERT_OBJECT_TYPE_POD)
	if err != nil {
		return err
	}

	for _, alert := range relatedAlerts {
		if alert.Status != model.ALERT_STATE_RESOLVED {
			err = a.store.UpdateAlertStatus(alert.ID, model.ALERT_STATE_RESOLVED)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a AlertStateManager) TrackPods(pods []*api.Pod) error {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		relatedAlerts, err := a.relatedAlerts(podName, model.ALERT_OBJECT_TYPE_POD)
		if err != nil {
			return err
		}

		unresolvedRelatedAlert := unresolvedRelatedAlert(relatedAlerts)

		if isPodErrorState(pod.Status) && isTrackedStatus(pod.Status) {
			if unresolvedRelatedAlert == nil {
				err := a.store.CreateAlert(&model.Alert{
					ObjectType:       model.ALERT_OBJECT_TYPE_POD,
					ObjectName:       podName,
					ObjectStatus:     pod.Status,
					ObjectStatusDesc: pod.StatusDescription,
					Status:           model.ALERT_STATE_PENDING,
					LastStateChange:  time.Now().Unix(),
				})
				if err != nil {
					return err
				}
			}
		} else if unresolvedRelatedAlert != nil {
			switch unresolvedRelatedAlert.ObjectStatus {
			case ImagePullBackOff:
				if !isPodStartupState(pod.Status) &&
					pod.Status != ImagePullBackOff &&
					pod.Status != ErrImagePull {
					err = a.resolve(unresolvedRelatedAlert)
					if err != nil {
						logrus.Warnf("could not resolve alert: %s", err)
					}
				}
			case CreateContainerConfigError:
				if !isPodStartupState(pod.Status) && pod.Status != CreateContainerConfigError {
					err = a.resolve(unresolvedRelatedAlert)
					if err != nil {
						logrus.Warnf("could not resolve alert: %s", err)
					}
				}
			case CrashLoopBackOff:
				if !isPodStartupState(pod.Status) {
					err = a.resolve(unresolvedRelatedAlert)
					if err != nil {
						logrus.Warnf("could not resolve alert: %s", err)
					}
				}
			}
		}
	}

	return nil
}

func (a AlertStateManager) resolve(unresolvedRelatedAlert *model.Alert) error {
	unresolvedRelatedAlert.Status = model.ALERT_STATE_RESOLVED

	err := a.store.UpdateAlertStatus(unresolvedRelatedAlert.ID, unresolvedRelatedAlert.Status)
	if err != nil {
		return err
	}

	msg := notifications.MessageFromAlert(*unresolvedRelatedAlert)
	a.notifManager.Broadcast(msg)

	return nil
}

func unresolvedRelatedAlert(alerts []*model.Alert) *model.Alert {
	for _, alert := range alerts {
		if alert.Status != model.ALERT_STATE_RESOLVED {
			return alert
		}
	}

	return nil
}

func isPodErrorState(state string) bool {
	return state != "Running" && state != "Pending" && state != "Terminating" &&
		state != "Succeeded" && state != "Unknown" && state != "ContainerCreating" &&
		state != "PodInitializing"
}

func isPodStartupState(state string) bool {
	return state == "PodInitializing" || state == "ContainerCreating"
}

func (a AlertStateManager) TrackEvents(events []api.Event) error {
	// if 1 == 1 {
	// 	return nil
	// }
	// for _, event := range events {
	// 	relatedObjectName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
	// 	deploymentName := fmt.Sprintf("%s/%s", event.Namespace, event.DeploymentName)
	// 	// fmt.Printf("%s - %s X %d\n\t%s\n", relatedObjectName, event.Status, event.Count, event.StatusDesc)
	// 	relatedAlert, err := a.relatedAlert(relatedObjectName, model.ALERT_OBJECT_TYPE_EVENT)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if a.statusNotChanged(relatedObjectName, event.Status) {
	// 		continue
	// 	}

	// 	if relatedAlert != nil && relatedAlert.IsFiring() {
	// 		continue
	// 	}

	// 	err = a.store.SaveOrUpdateKubeEvent(&model.KubeEvent{
	// 		Name:       relatedObjectName,
	// 		Status:     event.Status,
	// 		StatusDesc: event.StatusDesc,
	// 		Count:      int(event.Count),
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if relatedAlert != nil {
	// 		relatedAlert.Status = model.ALERT_STATE_PENDING
	// 		err = a.store.UpdateAlert(relatedAlert)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		err := a.store.CreateAlert(&model.Alert{
	// 			Type:           model.ALERT_OBJECT_TYPE_EVENT,
	// 			Name:           relatedObjectName,
	// 			DeploymentName: deploymentName,
	// 			Status:         model.ALERT_STATE_PENDING,
	// 			// StatusDesc:      event.StatusDesc,
	// 			LastStateChange: time.Now().Unix(),
	// 		})
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}

func (a AlertStateManager) relatedAlerts(objectName string, objectType string) ([]*model.Alert, error) {
	alerts, err := a.store.Alerts(objectName, objectType)
	if err != nil {
		return nil, fmt.Errorf("couldn't get alerts for pod from db: %s", err)
	}
	return alerts, nil
}

func isTrackedStatus(status string) bool {
	_, exists := thresholds[status]
	return exists
}
