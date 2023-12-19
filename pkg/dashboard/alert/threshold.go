package alert

import (
	"reflect"
	"time"

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
			waitTime:      120,
			waitToResolve: 300,
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
		"OOMKilled": oomKilledThreshold{
			waitToResolve: 300,
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
	waitTime      time.Duration
	waitToResolve time.Duration
}

type createContainerConfigErrorThreshold struct {
	waitTime time.Duration
}

type pendingThreshold struct {
	waitTime time.Duration
}

type oomKilledThreshold struct {
	waitToResolve time.Duration
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
	// event := relatedObject.(*api.Event)
	// alertPendingSinceInMinutes := time.Since(time.Unix(alert.PendingAt, 0)).Minutes()
	// countPerMinute := float64(event.Count) / alertPendingSinceInMinutes

	// return countPerMinute >= s.minimumCountPerMinute && event.Count >= s.minimumCount
	return false
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
	if pod.RunningSince == 0 {
		return false
	}
	if pod.Status != model.POD_RUNNING {
		return false
	}

	runningSince := time.Unix(pod.RunningSince, 0)
	waitToResolveTime := time.Now().Add(-time.Second * s.waitToResolve)
	return pod.Status == model.POD_RUNNING && runningSince.Before(waitToResolveTime)
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

func (s oomKilledThreshold) Reached(relatedObject interface{}, alert *model.Alert) bool {
	return true
}

func (s oomKilledThreshold) Resolved(relatedObject interface{}) bool {
	pod := relatedObject.(*model.Pod)
	if pod.RunningSince == 0 {
		return false
	}
	if pod.Status != model.POD_RUNNING {
		return false
	}

	runningSince := time.Unix(pod.RunningSince, 0)
	waitToResolveTime := time.Now().Add(-time.Second * s.waitToResolve)
	return pod.Status == model.POD_RUNNING && runningSince.Before(waitToResolveTime)
}

func (t imagePullBackOffThreshold) Text() string {
	return `
### When It Happens

ImagePullBackOff and ErrImagePull errors occur when Kubernetes cannot fetch the container image specified in your pod configuration.

### How to Fix It

You need to verify the correctness of your image name and double-check your image registry credentials.

- Run ` + "`" + `kubectl describe pod <pod-name>` + "`" + ` to cross check the image name.
- Check the exact error message at the bottom of the ` + "`" + `kubectl describe output` + "`" + `. It may have further clues.

If the image name is correct, check the access credentials you use with ` + "`" + `kubectl get pod <pod-name> -o=jsonpath='{.spec.imagePullSecrets[0].name}{"\n"}'` + "`" + ` then check the secret values with ` + "`" + `kubectl get secret <your-pull-secret> -o yaml` + "`" + `. You may feed the base64 encoded fields to ` + "`" + `echo xxx | base64 -d` + "`" + `
`
}

func (t crashLoopBackOffThreshold) Text() string {
	return `
### When It Happens

CrashLoopBackOff signifies that your application keeps starting up and then dying for some reason.

### How to Fix It

Investigate your application logs for bugs, misconfigurations, or resource issues. ` + "`" + `kubectl logs <pod-name>` + "`" + ` is your best bet. The ` + "`" + `--previous` + "`" + ` flag will dump pod logs (stdout) for a previous instantiation of the pod.
`
}

func (t createContainerConfigErrorThreshold) Text() string {
	return `
### When It Happens

These errors crop up when Kubernetes encounters problems creating containers: a misconfigured ConfigMap or Secret is the most common reason.

### How to Fix It

Run ` + "`" + `kubectl describe pod <pod-name>` + "`" + ` and check the error message at the bottom of the output. It will highlight if you misspelled a ConfigMap name, or a Secret is not created yet.

Remember, if you don't see error messages at the end of ` + "`" + `kubectl describe` + "`" + `, restart the pod by deleting it. Error events are only visible for one hour after pod start.
`
}

func (t pendingThreshold) Text() string {
	return `
	### When It Happens

A pending pod is a pod that Kubernetes can't schedule on a node, often due to resource constraints or node troubles.

### How to Fix It

Dive into the events section of your pod's description using ` + "`" + `kubectl describe` + "`" + ` to spot scheduling issues.

Verify that your cluster has enough resources available by ` + "`" + `kubectl describe node <node-x>` + "`" + `	
`
}

func (t failedEventThreshold) Text() string {
	return "TODO"
}

func (t oomKilledThreshold) Text() string {
	return `
### When It Happens

Running out of memory can lead to your pod's restart. Sadly the OOMKilled error is not easy to spot.
	
### How to Fix It
	
You need a monitoring solution to chart your pod's memory usage over time. If your pod is reaching the resource limits in your pod specification, Kubernetes will restart your pod.
	
Correlate your restart times with your pod memory usage to confirm the out of memory situation and adjust your pod resource limits accordingly.

You can also use the ` + "`" + `kubectl describe pod <pod-name>` + "`" + ` command and look for the ` + "`" + `Last State` + "`" + ` section to confirm that indeed it is the lack of memory that restarted the pod.
`
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

func (t oomKilledThreshold) Name() string {
	return "OOMKilled"
}
