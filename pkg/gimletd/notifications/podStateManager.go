package notifications

import (
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/sirupsen/logrus"
)

type podStateManager struct {
	notifManager Manager
	store        store.Store
	waitTime     time.Duration
}

func NewPodStateManager(notifManager Manager, store store.Store, waitTime time.Duration) *podStateManager {
	return &podStateManager{notifManager: notifManager, store: store, waitTime: waitTime}
}

func (p podStateManager) Track(pods []api.Pod) {
	p.trackStates(pods)
}

func (p podStateManager) trackStates(pods []api.Pod) {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		currentTime := time.Now().Unix()

		if podErrorState(pod.Status) {
			err := p.store.SaveOrUpdatePod(&model.Pod{
				Name:                podName,
				Status:              pod.Status,
				StatusDesc:          pod.StatusDescription,
				AlertState:          "Pending",
				AlertStateTimestamp: currentTime,
			})
			if err != nil {
				logrus.Errorf("could't save or update pod: %s", err)
				continue
			}
		} else {
			err := p.store.SaveOrUpdatePod(&model.Pod{
				Name:                podName,
				Status:              pod.Status,
				StatusDesc:          pod.StatusDescription,
				AlertState:          "OK",
				AlertStateTimestamp: currentTime,
			})
			if err != nil {
				logrus.Errorf("could't save or update pod: %s", err)
				continue
			}
		}
	}
}

func (p podStateManager) NotificationManager() {
	for {
		pods, err := p.store.Pods()
		if err != nil {
			logrus.Errorf("could't get pods from db: %s", err)
		}

		for _, pod := range pods {
			if pod.AlertState == "Pending" && p.isPendingStateExpired(pod.AlertStateTimestamp) {
				// p.notifManager.Broadcast(msg)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (p podStateManager) isPendingStateExpired(alertTimestamp int64) bool {
	podAlertTime := time.Unix(alertTimestamp, 0)
	managerWaitTime := time.Now().Add(-time.Minute * p.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
