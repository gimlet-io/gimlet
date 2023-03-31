package alert

import (
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	isReached(relatedObject interface{}, alert *model.Alert) bool
}

type podThreshold struct {
	waitTime time.Duration
}

type eventThreshold struct {
	expectedCountPerMinute float64
	expectedCount          int
}

type noopThreshold struct {
}

func (p podThreshold) isReached(relatedObject interface{}, alert *model.Alert) bool {
	podLastStateChangeTime := time.Unix(alert.LastStateChange, 0)
	waitTime := time.Now().Add(-time.Minute * p.waitTime)

	return podLastStateChangeTime.Before(waitTime)
}

func (e eventThreshold) isReached(relatedObject interface{}, alert *model.Alert) bool {
	event := relatedObject.(*model.KubeEvent)
	lastStateChangeInMinutes := time.Since(time.Unix(alert.LastStateChange, 0)).Minutes()
	countPerMinute := float64(event.Count) / lastStateChangeInMinutes

	return countPerMinute >= e.expectedCountPerMinute && event.Count >= e.expectedCount
}

func (n noopThreshold) isReached(relatedObject interface{}, alert *model.Alert) bool {
	return false
}
