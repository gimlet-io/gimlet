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

// TODO check if can be more good to look at it
func toThreshold(a *model.Alert) threshold {
	if a.Type == "pod" {
		return &podStrategy{
			pod:      *a,
			waitTime: 1, // TODO make it configurable
		}
	}

	return &eventStrategy{
		event:                  *a,
		expectedCountPerMinute: 1, // TODO make it configurable
		expectedCount:          6, // TODO make it configurable
	}
}
