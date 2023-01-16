package notifications

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/sirupsen/logrus"
)

type PodStateManager struct {
	notifManager Manager
	store        store.Store
	waitTime     time.Duration
}

func NewPodStateManager(notifManager Manager, store store.Store, waitTime time.Duration) *PodStateManager {
	return &PodStateManager{notifManager: notifManager, store: store, waitTime: waitTime}
}

func (p PodStateManager) Track(pods []*api.Pod) {
	p.trackStates(pods)
}

func (p PodStateManager) Run() {
	for {
		p.setFiringState()

		time.Sleep(p.waitTime * time.Minute)
	}
}

func (p PodStateManager) trackStates(pods []*api.Pod) {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		currentTime := time.Now().Unix()

		if p.alreadyAlerted(podName) || p.statusNotChanged(podName, pod.Status) {
			continue
		}

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
			}
		}
	}
}

func (p PodStateManager) setFiringState() {
	pods, err := p.store.Pods()
	if err != nil {
		logrus.Errorf("could't get pods from db: %s", err)
	}

	for _, pod := range pods {
		if pod.AlertState == "Pending" && p.waiTimeIsSoonerThan(pod.AlertStateTimestamp) {
			msg := MessageFromFailedPod(*pod)
			p.notifManager.Broadcast(msg)

			err := p.store.SaveOrUpdatePod(&model.Pod{
				Name:                pod.Name,
				Status:              pod.Status,
				StatusDesc:          pod.StatusDesc,
				AlertState:          "Firing",
				AlertStateTimestamp: pod.AlertStateTimestamp,
			})
			if err != nil {
				logrus.Errorf("could't save or update pod: %s", err)
			}
		}
	}
}

func (p PodStateManager) waiTimeIsSoonerThan(alertTimestamp int64) bool {
	podAlertTime := time.Unix(alertTimestamp, 0)
	managerWaitTime := time.Now().Add(-time.Minute * p.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func (p PodStateManager) alreadyAlerted(podName string) bool {
	pod, err := p.store.Pod(podName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return pod.AlertState == "Firing"
}

func (p PodStateManager) statusNotChanged(podName string, podStatus string) bool {
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
