package alert

import (
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	isFired() bool
	toAlert() model.Alert
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
	podAlertTime := time.Unix(s.pod.LastStateChange, 0)
	managerWaitTime := time.Now().Add(-time.Minute * s.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func (s eventStrategy) isFired() bool {
	firstTimestampSinceInMinutes := time.Since(time.Unix(s.event.LastStateChange, 0)).Minutes()
	countPerMinute := float64(s.event.Count) / firstTimestampSinceInMinutes

	return countPerMinute >= s.expectedCountPerMinute && s.event.Count >= s.expectedCount
}

func (s podStrategy) toAlert() model.Alert {
	return s.pod
}

func (s eventStrategy) toAlert() model.Alert {
	return s.event
}

func podTypeToThreshold(a *model.Alert, waitTime time.Duration) threshold {
	return &podStrategy{
		pod:      *a,
		waitTime: waitTime,
	}
}
func eventTypeToThreshold(a *model.Alert, expectedCount int32, expectedCountPerMinute float64) threshold {
	return &eventStrategy{
		event:                  *a,
		expectedCount:          expectedCount,
		expectedCountPerMinute: expectedCountPerMinute,
	}
}
