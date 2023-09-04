// Copyright 2019 Laszlo Fogas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const AnnotationGitRepository = "gimlet.io/git-repository"
const AnnotationGitSha = "gimlet.io/git-sha"

type KubeEnv struct {
	Name          string
	Namespace     string
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
}

func (e *KubeEnv) Services(repo string) ([]*api.Stack, error) {
	annotatedServices, err := e.annotatedServices(repo)
	if err != nil {
		logrus.Errorf("could not get 1 %v", err)
		return nil, err
	}

	d, err := e.Client.AppsV1().Deployments(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get deployments: %s", err)
	}

	i, err := e.Client.NetworkingV1().Ingresses(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get ingresses: %s", err)
	}

	var stacks []*api.Stack
	for _, service := range annotatedServices {
		deployment, err := e.deploymentForService(service, d.Items)
		if err != nil {
			return nil, fmt.Errorf("could not get deployment for service: %s", err)
		}

		var ingresses []*api.Ingress
		for _, ingress := range i.Items {
			for _, rule := range ingress.Spec.Rules {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service.Name == service.Name {
						ingresses = append(ingresses, &api.Ingress{Name: ingress.Name, Namespace: ingress.Namespace, URL: rule.Host})
					}
				}
			}
		}

		stacks = append(stacks, &api.Stack{
			Repo:       service.ObjectMeta.GetAnnotations()[AnnotationGitRepository],
			Service:    &api.Service{Name: service.Name, Namespace: service.Namespace},
			Deployment: deployment,
			Ingresses:  ingresses,
		})
	}

	return stacks, nil
}

var gitRepositoryResource = schema.GroupVersionResource{
	Group:    "source.toolkit.fluxcd.io",
	Version:  "v1beta1",
	Resource: "gitrepositories",
}

var kustomizationResource = schema.GroupVersionResource{
	Group:    "kustomize.toolkit.fluxcd.io",
	Version:  "v1beta1",
	Resource: "kustomizations",
}

var helmReleaseResource = schema.GroupVersionResource{
	Group:    "helm.toolkit.fluxcd.io",
	Version:  "v2beta1",
	Resource: "helmreleases",
}

func (e *KubeEnv) GitRepositories() ([]*api.GitRepository, error) {
	gitRepositories, err := e.DynamicClient.
		Resource(gitRepositoryResource).
		Namespace("").
		List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := []*api.GitRepository{}
	for _, g := range gitRepositories.Items {
		gitRepository, err := asGitRepository(g)
		if err != nil {
			logrus.Warnf("could not parse gitrepository: %s", err)
		}
		result = append(result, gitRepository)
	}

	return result, nil
}

func (e *KubeEnv) Kustomizations() ([]*api.Kustomization, error) {
	kustomizations, err := e.DynamicClient.
		Resource(kustomizationResource).
		Namespace("").
		List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := []*api.Kustomization{}
	for _, k := range kustomizations.Items {
		kustomization, err := asKustomization(k)
		if err != nil {
			logrus.Warnf("could not parse kustomization: %s", err)
		}
		result = append(result, kustomization)
	}

	return result, nil
}

func (e *KubeEnv) HelmReleases() ([]*api.HelmRelease, error) {
	helmReleases, err := e.DynamicClient.
		Resource(helmReleaseResource).
		Namespace("").
		List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := []*api.HelmRelease{}
	for _, h := range helmReleases.Items {
		helmRelease, err := asHelmRelease(h)
		if err != nil {
			logrus.Warnf("could not parse kustomization: %s", err)
		}
		result = append(result, helmRelease)
	}

	return result, nil
}

func statusAndMessage(conditions []interface{}) (string, string, int64) {
	if c := findStatusCondition(conditions, meta.ReadyCondition); c != nil {
		transitionTime, _ := time.Parse(time.RFC3339, c["lastTransitionTime"].(string))
		return c["reason"].(string), c["message"].(string), transitionTime.Unix()
	}
	return string(metav1.ConditionFalse), "waiting to be reconciled", time.Now().Unix()
}

// findStatusCondition finds the conditionType in conditions.
func findStatusCondition(conditions []interface{}, conditionType string) map[string]interface{} {
	for _, c := range conditions {
		cMap := c.(map[string]interface{})
		if cMap["type"] == conditionType {
			return cMap
		}
	}

	return nil
}

func asGitRepository(g unstructured.Unstructured) (*api.GitRepository, error) {
	statusMap := g.Object["status"].(map[string]interface{})

	artifact, ok := statusMap["artifact"].(map[string]interface{})
	if !ok {
		// TODO handle case
	}
	revision := artifact["revision"].(string)

	conditions, ok := statusMap["conditions"].([]interface{})
	if !ok {
		// TODO handle case
	}

	status, statusDesc, lastTransitionTime := statusAndMessage(conditions)

	return &api.GitRepository{
		Name:               g.GetName(),
		Namespace:          g.GetNamespace(),
		Revision:           revision,
		LastTransitionTime: lastTransitionTime,
		Status:             status,
		StatusDesc:         statusDesc,
	}, nil
}

func asKustomization(g unstructured.Unstructured) (*api.Kustomization, error) {
	statusMap := g.Object["status"].(map[string]interface{})

	spec, ok := g.Object["spec"].(map[string]interface{})
	if !ok {
		// TODO handle case
	}
	path := spec["path"].(string)
	prune := spec["prune"].(bool)
	sourceRef := spec["sourceRef"].(map[string]interface{})
	gitRepository := sourceRef["name"].(string)

	conditions, ok := statusMap["conditions"].([]interface{})
	if !ok {
		// TODO handle case
	}

	status, statusDesc, lastTransitionTime := statusAndMessage(conditions)

	return &api.Kustomization{
		Name:               g.GetName(),
		Namespace:          g.GetNamespace(),
		GitRepository:      gitRepository,
		Path:               path,
		Prune:              prune,
		LastTransitionTime: lastTransitionTime,
		Status:             status,
		StatusDesc:         statusDesc,
	}, nil
}

func asHelmRelease(g unstructured.Unstructured) (*api.HelmRelease, error) {
	statusMap := g.Object["status"].(map[string]interface{})

	conditions, ok := statusMap["conditions"].([]interface{})
	if !ok {
		// TODO handle case
	}

	status, statusDesc, lastTransitionTime := statusAndMessage(conditions)

	return &api.HelmRelease{
		Name:               g.GetName(),
		Namespace:          g.GetNamespace(),
		LastTransitionTime: lastTransitionTime,
		Status:             status,
		StatusDesc:         statusDesc,
	}, nil
}

func (e *KubeEnv) WarningEvents(repo string) ([]api.Event, error) {
	integratedServices, err := e.annotatedServices(repo)
	if err != nil {
		return nil, err
	}

	allDeployments, err := e.Client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	allPods, err := e.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	events, err := e.Client.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var typeWarningEvents []api.Event
	for _, svc := range integratedServices {
		for _, deployment := range allDeployments.Items {
			if SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
				for _, pod := range allPods.Items {
					if HasLabels(deployment.Spec.Selector.MatchLabels, pod.GetObjectMeta().GetLabels()) &&
						pod.Namespace == deployment.Namespace {
						for _, event := range events.Items {
							if event.Type == v1.EventTypeWarning && (event.InvolvedObject.Name == pod.Name || event.InvolvedObject.Name == deployment.Name) {
								typeWarningEvents = append(typeWarningEvents, api.Event{
									FirstTimestamp: event.FirstTimestamp.Unix(),
									Count:          count(event),
									Namespace:      event.Namespace,
									Name:           event.InvolvedObject.Name,
									DeploymentName: deployment.Name,
									Status:         event.Reason,
									StatusDesc:     event.Message,
								})
							}
						}
					}
				}
			}
		}
	}
	return typeWarningEvents, nil
}

// annotatedServices returns all services that are enabled for Gimlet
func (e *KubeEnv) annotatedServices(repo string) ([]v1.Service, error) {
	svc, err := e.Client.CoreV1().Services(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []v1.Service
	for _, s := range svc.Items {
		if _, ok := s.ObjectMeta.GetAnnotations()[AnnotationGitRepository]; ok {
			if repo == "" {
				services = append(services, s)
			} else if repo == s.ObjectMeta.GetAnnotations()[AnnotationGitRepository] {
				services = append(services, s)
			}
		}
	}

	return services, nil
}

func (e *KubeEnv) deploymentForService(service v1.Service, deployments []appsv1.Deployment) (*api.Deployment, error) {
	var deployment *api.Deployment

	for _, d := range deployments {
		if SelectorsMatch(d.Spec.Selector.MatchLabels, service.Spec.Selector) {
			var sha string
			if hash, ok := d.GetAnnotations()[AnnotationGitSha]; ok {
				sha = hash
			}

			var pods []*api.Pod
			set := labels.Set(service.Spec.Selector)
			p, err := e.Client.CoreV1().Pods(e.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.AsSelector().String()})
			if err != nil {
				return nil, err
			}
			for _, pod := range p.Items {
				podStatus := podStatus(pod)
				podLogs := ""
				if "CrashLoopBackOff" == podStatus || "Error" == podStatus {
					podLogs = logs(e, pod)
				}
				pods = append(pods, &api.Pod{Name: pod.Name, DeploymentName: d.Name, Namespace: pod.Namespace, Status: podStatus, StatusDescription: podErrorCause(pod), Logs: podLogs})
			}

			deployment = &api.Deployment{Name: d.Name, Namespace: d.Namespace, Pods: pods, SHA: sha}
		}
	}

	return deployment, nil
}

func logs(e *KubeEnv, pod v1.Pod) string {
	podLogs := ""
	req := e.Client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{TailLines: fifty()})
	result := req.Do(context.TODO())
	if result.Error() != nil {
		logrus.Warnf("could not get logs %s", result.Error())
	} else {
		logBytes, err := result.Raw()
		if err != nil {
			logrus.Warnf("could not get logs %s", err.Error())
		} else {
			podLogs = string(logBytes)
		}
	}
	return podLogs
}

func fifty() *int64 {
	fifty := int64(50)
	return &fifty
}

func podErrorCause(pod v1.Pod) string {
	if v1.PodPending == pod.Status.Phase ||
		v1.PodRunning == pod.Status.Phase {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				return fmt.Sprint(containerStatus.State.Waiting.Message)
			}
		}
	}

	return ""
}

func podStatus(pod v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating" //https://github.com/kubernetes/kubernetes/issues/61376#issuecomment-374437926
	}

	if v1.PodPending == pod.Status.Phase ||
		v1.PodRunning == pod.Status.Phase {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				return fmt.Sprint(containerStatus.State.Waiting.Reason)
			}
		}
	}

	return fmt.Sprint(pod.Status.Phase)
}

func SelectorsMatch(first map[string]string, second map[string]string) bool {
	if len(first) != len(second) {
		return false
	}

	for k, v := range first {
		if v2, ok := second[k]; ok {
			if v != v2 {
				return false
			}
		} else {
			return false
		}
	}

	for k2, v2 := range second {
		if v, ok := first[k2]; ok {
			if v2 != v {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func httpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   20 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}
}

func count(e v1.Event) int32 {
	if e.Series != nil {
		return e.Series.Count
	} else if e.Count > 1 {
		return e.Count
	}
	return 0
}
