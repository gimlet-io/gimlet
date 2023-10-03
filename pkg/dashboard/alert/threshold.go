package alert

import (
	"reflect"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
)

type threshold interface {
	Reached(relatedObject interface{}, alert *model.Alert) bool
	Resolved(relatedObject interface{}) bool
	Text() string
	Name() string
}

func Thresholds() map[string]threshold {
	return map[string]threshold{
		"ImagePullBackOff": imagePullBackOffThreshold{
			waitTime: 120,
		},
		"CrashLoopBackOff": crashLoopBackOffThreshold{
			waitTime: 120,
		},
		"CreateContainerConfigError": createContainerConfigErrorThreshold{
			waitTime: 60,
		},
		"Pending": pendingThreshold{
			waitTime: 600,
		},
		"Failed": failedEventThreshold{
			minimumCount:          6,
			minimumCountPerMinute: 1,
		},
	}
}

func ThresholdByType(thresholds map[string]threshold, thresholdTypeString string) threshold {
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

type imagePullBackOffThreshold struct {
	waitTime time.Duration
}

type failedEventThreshold struct {
	minimumCount          int32
	minimumCountPerMinute float64
}

type crashLoopBackOffThreshold struct {
	waitTime time.Duration
}

type createContainerConfigErrorThreshold struct {
	waitTime time.Duration
}

type pendingThreshold struct {
	waitTime time.Duration
}

func (s imagePullBackOffThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	alertPendingSince := time.Unix(alert.PendingAt, 0)
	waitTime := time.Now().Add(-time.Second * s.waitTime)
	return alertPendingSince.Before(waitTime)
}

func (s imagePullBackOffThreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	return pod.Status == model.POD_RUNNING
}

func (s failedEventThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	event := relatedObject.(*api.Event)
	alertPendingSinceInMinutes := time.Since(time.Unix(alert.PendingAt, 0)).Minutes()
	countPerMinute := float64(event.Count) / alertPendingSinceInMinutes

	return countPerMinute >= s.minimumCountPerMinute && event.Count >= s.minimumCount
}

func (s failedEventThreshold) Resolved(relatedObject interface{}) bool {
	return false
}

func (s crashLoopBackOffThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	alertPendingSince := time.Unix(alert.PendingAt, 0)
	waitTime := time.Now().Add(-time.Second * s.waitTime)
	return alertPendingSince.Before(waitTime)
}

func (s crashLoopBackOffThreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	return pod.Status == model.POD_RUNNING
}

func (s createContainerConfigErrorThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	alertPendingSince := time.Unix(alert.PendingAt, 0)
	waitTime := time.Now().Add(-time.Second * s.waitTime)
	return alertPendingSince.Before(waitTime)
}

func (s createContainerConfigErrorThreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	return pod.Status == model.POD_RUNNING
}

func (s pendingThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	alertPendingSince := time.Unix(alert.PendingAt, 0)
	waitTime := time.Now().Add(-time.Second * s.waitTime)
	return alertPendingSince.Before(waitTime)
}

func (s pendingThreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	return pod.Status != model.POD_PENDING
}

func (t imagePullBackOffThreshold) Text() string {
	return "TODO"
}

func (t crashLoopBackOffThreshold) Text() string {
	return "TODO"
}

func (t createContainerConfigErrorThreshold) Text() string {
	return "TODO"
}

func (t pendingThreshold) Text() string {
	return "TODO"
}

func (t failedEventThreshold) Text() string {
	return "TODO"
}

func (t imagePullBackOffThreshold) Name() string {
	return "ImagePullBackOff"
}

func (t crashLoopBackOffThreshold) Name() string {
	return "crashLoopBackOff"
}

func (t createContainerConfigErrorThreshold) Name() string {
	return "CreateContainerConfigError"
}

func (t pendingThreshold) Name() string {
	return "Pending"
}

func (t failedEventThreshold) Name() string {
	return "TODO"
}
