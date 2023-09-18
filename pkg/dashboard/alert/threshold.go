package alert

import (
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	Reached(relatedObject interface{}, alert *model.Alert) bool
}

func tresholds() map[string]threshold {
	return map[string]threshold{
		"ImagePullBackOff": imagePullBackOffTreshold{
			waitTime: 2,
		},
		"Failed": failedEventThreshold{
			MinimumCount:   6,
			CountPerMinute: 1,
		},
	}
}

type imagePullBackOffTreshold struct {
	waitTime time.Duration
}

type failedEventThreshold struct {
	MinimumCount   int32
	CountPerMinute float64
}

func (s imagePullBackOffTreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	alertPendingSince := time.Unix(alert.LastStateChange, 0)
	waitTime := time.Now().Add(-time.Minute * s.waitTime)
	return alertPendingSince.Before(waitTime)
}

func (s failedEventThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	lastStateChangeInMinutes := time.Since(time.Unix(alert.LastStateChange, 0)).Minutes()
	countPerMinute := float64(alert.Count) / lastStateChangeInMinutes

	return countPerMinute >= s.CountPerMinute && alert.Count >= s.MinimumCount
}
