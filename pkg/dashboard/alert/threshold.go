package alert

import (
	"reflect"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	// Candidate(relatedObject interface{}) bool
	Reached(relatedObject interface{}, alert *model.Alert) bool
	Resolved(relatedObject interface{}) bool
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

func thresholdByType(thresholds map[string]threshold, thresholdTypeString string) threshold {
	for _, t := range thresholds {
		if thresholdType(t) == thresholdTypeString {
			return t
		}
	}
	return nil
}

func thresholdType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
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

func (s imagePullBackOffTreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	return pod.Status == "Running"
}

func (s failedEventThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	lastStateChangeInMinutes := time.Since(time.Unix(alert.LastStateChange, 0)).Minutes()
	countPerMinute := float64(alert.Count) / lastStateChangeInMinutes

	return countPerMinute >= s.CountPerMinute && alert.Count >= s.MinimumCount
}

func (s failedEventThreshold) Resolved(relatedObject interface{}) bool {
	return false
}
