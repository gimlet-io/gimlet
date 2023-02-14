package server

import (
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	isFired() bool
	model() model.Alert
}

type podStrategy struct {
	pod      model.Alert
	waitTime time.Duration
}

type eventStrategy struct {
	event                  model.Alert
	expectedCountPerMinute float64
	expectedCount          int32
}

func (s podStrategy) isFired() bool {
	podAlertTime := time.Unix(s.pod.Fired, 0)
	managerWaitTime := time.Now().Add(-time.Minute * s.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func (s eventStrategy) isFired() bool {
	firstTimestampSinceInMinutes := time.Since(time.Unix(s.event.Fired, 0)).Minutes()
	countPerMinute := float64(s.event.Count) / firstTimestampSinceInMinutes

	return countPerMinute >= s.expectedCountPerMinute && s.event.Count >= s.expectedCount
}

func (s podStrategy) model() model.Alert {
	return s.pod
}

func (s eventStrategy) model() model.Alert {
	return s.event
}

func thresholdFromPod(pod model.Alert, waitTime time.Duration) threshold {
	return &podStrategy{
		pod:      pod,
		waitTime: waitTime,
	}
}

func thresholdFromEvent(event model.Alert) threshold {
	return &eventStrategy{
		event:                  event,
		expectedCountPerMinute: 1, // TODO make it configurable
		expectedCount:          6, // TODO make it configurable
	}
}
