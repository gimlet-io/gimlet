package notifications

import (
	"database/sql"
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
	waitTime     int
}

func NewPodStateManager(notifManager Manager, store store.Store, waitTime int) *podStateManager {
	return &podStateManager{notifManager: notifManager, store: store, waitTime: waitTime}
}

func (p podStateManager) Track(pods []api.Pod) {
	p.trackStates(pods, p.store)
}

func (p podStateManager) trackStates(pods []api.Pod, store store.Store) {
	for _, pod := range pods {
		deployment := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podFromStore, err := store.Pod(deployment)
		if err == sql.ErrNoRows {
			err = store.SaveOrUpdatePod(&model.Pod{
				Name:       fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				Status:     pod.Status,
				StatusDesc: pod.StatusDescription,
			})
			if err != nil {
				logrus.Errorf("couldn't save or update pod: %s", err)
				continue
			}
			if podErrorState(pod.Status) {
				go p.checkWithDelay(store, pod)
			}
			continue
		} else if err != nil {
			logrus.Errorf("couldn't get pod from db: %s", err)
			continue
		}

		if podErrorState(pod.Status) && pod.Status != podFromStore.Status {
			err = store.SaveOrUpdatePod(&model.Pod{
				Name:       fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				Status:     pod.Status,
				StatusDesc: pod.StatusDescription,
			})
			if err != nil {
				logrus.Errorf("couldn't save or update pod: %s", err)
				continue
			}
			go p.checkWithDelay(store, pod)
		}
	}
}

func (p podStateManager) checkWithDelay(store store.Store, pod api.Pod) {
	time.Sleep(time.Duration(p.waitTime) * time.Minute)

	podFromStore, err := store.Pod(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("couldn't get pod from db: %s", err)
	}

	if podErrorState(podFromStore.Status) {
		//TODO send out notification
		// p.notifManager.Broadcast(msg)
	}
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
