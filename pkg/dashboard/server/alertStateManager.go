package server

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

type alertStateManager struct {
	notifManager notifications.Manager
	store        store.Store
	waitTime     time.Duration
}

func NewAlertStateManager(notifManager notifications.Manager, store store.Store, waitTime time.Duration) *alertStateManager {
	return &alertStateManager{notifManager: notifManager, store: store, waitTime: waitTime}
}

func (p alertStateManager) Track(pods []*api.Pod) {
	p.trackStates(pods)
}

func (p alertStateManager) TrackEvents(events []api.Event) {
	p.trackEvents(events)
}

func (p alertStateManager) Run() {
	for {
		p.setFiringState()
		p.setFiringStateForEvents()
		time.Sleep(p.waitTime * time.Minute)
	}
}

func (p alertStateManager) Delete(podName string) {
	err := p.store.DeletePod(podName)
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("could't delete pod: %s", err)
	}
}

func (p alertStateManager) DeleteEvent(name string) {
	err := p.store.DeleteEvent(name)
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("could't delete event: %s", err)
	}
}

func (p alertStateManager) Alerts() ([]*model.Alert, error) {
	alerts, err := p.store.Alerts()
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

func (p alertStateManager) trackStates(pods []*api.Pod) {
	for _, pod := range pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		currentTime := time.Now().Unix()
		alertState := "Pending"

		if p.podAlreadyAlerted(podName) || p.statusNotChanged(podName, pod.Status) {
			continue
		}

		if !podErrorState(pod.Status) {
			alertState = "OK"
		}

		err := p.store.SaveOrUpdatePod(&model.Pod{
			Name:                podName,
			Status:              pod.Status,
			StatusDesc:          pod.StatusDescription,
			AlertState:          alertState,
			AlertStateTimestamp: currentTime,
		})
		if err != nil {
			logrus.Errorf("could't save or update pod: %s", err)
		}
	}
}

func (p alertStateManager) trackEvents(events []api.Event) {
	for _, event := range events {
		eventName := fmt.Sprintf("%s/%s", event.Namespace, event.Name)
		currentTime := time.Now().Unix()

		if p.eventAlreadyAlerted(eventName) || p.eventCountLessThanSaved(eventName, event.Count) {
			continue
		}

		err := p.store.SaveOrUpdateEvent(&model.Event{
			FirstTimestamp:      event.FirstTimestamp,
			Count:               event.Count,
			Name:                eventName,
			Status:              event.Status,
			StatusDesc:          event.StatusDesc,
			AlertState:          "Pending",
			AlertStateTimestamp: currentTime,
		})
		if err != nil {
			logrus.Errorf("could't save or update event: %s", err)
		}
	}
}

func (p alertStateManager) setFiringState() {
	pods, err := p.store.PendingPods()
	if err != nil {
		logrus.Errorf("could't get pods from db: %s", err)
	}

	for _, pod := range pods {
		if p.waiTimeIsSoonerThan(pod.AlertStateTimestamp) {
			msg := notifications.MessageFromFailedPod(*pod)
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

			err = p.store.SaveAlert(&model.Alert{
				Type:       "pod",
				Name:       pod.Name,
				Env:        "TODO", //TODO or deployment name
				Repo:       "TODO",
				Status:     pod.Status,
				StatusDesc: pod.StatusDesc,
				Fired:      time.Now().Unix(),
			})
			if err != nil {
				logrus.Errorf("could't save alert: %s", err)
			}
		}
	}
}

func (p alertStateManager) setFiringStateForEvents() {
	events, err := p.store.PendingEvents()
	if err != nil {
		logrus.Errorf("could't get pods from db: %s", err)
	}

	for _, event := range events {
		if p.countPerMinuteMoreThanOne(event.FirstTimestamp, event.Count) {
			msg := notifications.MessageFromWarningEvent(*event)
			p.notifManager.Broadcast(msg)

			err := p.store.SaveOrUpdateEvent(&model.Event{
				FirstTimestamp:      event.FirstTimestamp,
				Count:               event.Count,
				Name:                event.Name,
				Status:              event.Status,
				StatusDesc:          event.StatusDesc,
				AlertState:          "Firing",
				AlertStateTimestamp: event.AlertStateTimestamp,
			})
			if err != nil {
				logrus.Errorf("could't save or update event: %s", err)
			}

			err = p.store.SaveAlert(&model.Alert{
				Type:       "event",
				Name:       event.Name,
				Env:        "TODO", //TODO or deployment name
				Repo:       "TODO",
				Status:     event.Status,
				StatusDesc: event.StatusDesc,
				Fired:      time.Now().Unix(),
			})
			if err != nil {
				logrus.Errorf("could't save alert: %s", err)
			}
		}
	}
}

func (p alertStateManager) waiTimeIsSoonerThan(alertTimestamp int64) bool {
	podAlertTime := time.Unix(alertTimestamp, 0)
	managerWaitTime := time.Now().Add(-time.Minute * p.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func (p alertStateManager) podAlreadyAlerted(podName string) bool {
	pod, err := p.store.Pod(podName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return pod.AlertState == "Firing"
}

func (p alertStateManager) eventAlreadyAlerted(eventName string) bool {
	event, err := p.store.Event(eventName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return event.AlertState == "Firing"
}

func (p alertStateManager) eventCountLessThanSaved(eventName string, count int32) bool {
	event, err := p.store.Event(eventName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return event.Count > count
}

func (p alertStateManager) statusNotChanged(podName string, podStatus string) bool {
	podFromDb, err := p.store.Pod(podName)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		logrus.Errorf("could't get pod from db: %s", err)
		return false
	}

	return podStatus == podFromDb.Status
}

func (p alertStateManager) countPerMinuteMoreThanOne(firstTimestamp int64, count int32) bool {
	firstTimestampSinceInMinutes := time.Since(time.Unix(firstTimestamp, 0)).Minutes()
	countPerMinute := float64(count) / firstTimestampSinceInMinutes

	return countPerMinute >= 1 && count >= 6
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
