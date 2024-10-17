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

	"github.com/gimlet-io/capacitor/pkg/flux"
	"github.com/gimlet-io/gimlet/pkg/dashboard/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const AnnotationGitRepository = "gimlet.io/git-repository"
const AnnotationGitSha = "gimlet.io/git-sha"
const AnnotationGitBranch = "gimlet.io/git-branch"
const AnnotationDocsLink = "v1alpha1.opensca.dev/documentation"
const AnnotationLogsLink = "v1alpha1.opensca.dev/logs"
const AnnotationMetricsLink = "v1alpha1.opensca.dev/metrics"
const AnnotationTracesLink = "v1alpha1.opensca.dev/traces"
const AnnotationIssuesLink = "v1alpha1.opensca.dev/issues"
const AnnotationOwnerName = "v1alpha1.opensca.dev/owner.name"
const AnnotationOwnerIm = "v1alpha1.opensca.dev/owner.im"
const AnnotationBelongsTo = "gimlet.io/belongs-to"

type KubeEnv struct {
	Name          string
	Namespace     string
	Config        *rest.Config
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
	Perf          *prometheus.HistogramVec
}

func (e *KubeEnv) Services() ([]*api.Stack, error) {
	t0 := time.Now()
	annotatedServices, err := e.annotatedServices(AnnotationGitRepository)
	if err != nil {
		logrus.Errorf("could not get 1 %v", err)
		return nil, err
	}
	e.Perf.WithLabelValues("gimlet_agent_services").Observe(float64(time.Since(t0).Seconds()))

	t0 = time.Now()
	d, err := e.Client.AppsV1().Deployments(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get deployments: %s", err)
	}
	e.Perf.WithLabelValues("gimlet_agent_deployments").Observe(float64(time.Since(t0).Seconds()))

	t0 = time.Now()
	s, err := e.Client.AppsV1().StatefulSets(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get statefulsets: %s", err)
	}
	e.Perf.WithLabelValues("gimlet_agent_statefulsets").Observe(float64(time.Since(t0).Seconds()))

	t0 = time.Now()
	i, err := e.Client.NetworkingV1().Ingresses(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get ingresses: %s", err)
	}
	e.Perf.WithLabelValues("gimlet_agent_ingresses").Observe(float64(time.Since(t0).Seconds()))

	t0 = time.Now()
	pods, err := e.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get pods: %s", err)
	}
	e.Perf.WithLabelValues("gimlet_agent_pods").Observe(float64(time.Since(t0).Seconds()))

	t0 = time.Now()
	var stacks []*api.Stack
	for _, service := range annotatedServices {
		deployment, err := e.deploymentForService(service, d.Items, s.Items)
		if err != nil {
			return nil, fmt.Errorf("could not get deployment for service: %s", err)
		}

		if deployment != nil {
			deployment.Pods = []*api.Pod{}
			for _, pod := range pods.Items {
				if labelsMatchSelectors(pod.ObjectMeta.Labels, service.Spec.Selector) {
					podStatus := podStatus(pod)
					podLogs := ""
					if podStatus == "CrashLoopBackOff" || podStatus == "Error" {
						podLogs = logs(e, pod)
					}
					deployment.Pods = append(deployment.Pods, &api.Pod{Name: pod.Name, DeploymentName: deployment.Name, Namespace: pod.Namespace, Status: podStatus, StatusDescription: podErrorCause(pod), Logs: podLogs, ImChannelId: service.ObjectMeta.GetAnnotations()[AnnotationOwnerIm], Containers: PodContainers(pod.Spec)})
				}
			}
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
			Osca:       getOpenServiceCatalogAnnotations(service),
			Service:    &api.Service{Name: service.Name, Namespace: service.Namespace},
			Deployment: deployment,
			Ingresses:  ingresses,
		})
	}
	e.Perf.WithLabelValues("gimlet_agent_stacks").Observe(float64(time.Since(t0).Seconds()))

	dependencyServices, err := e.annotatedServices(AnnotationBelongsTo)
	if err != nil {
		logrus.Errorf("could not get 1 %v", err)
		return nil, err
	}
	e.Perf.WithLabelValues("gimlet_agent_dependency_services").Observe(float64(time.Since(t0).Seconds()))

	for _, service := range dependencyServices {
		deployment, err := e.deploymentForService(service, d.Items, s.Items)
		if err != nil {
			return nil, fmt.Errorf("could not get deployment for service: %s", err)
		}
		if deployment == nil {
			continue
		}

		if deployment != nil {
			deployment.Pods = []*api.Pod{}
			for _, pod := range pods.Items {
				if labelsMatchSelectors(pod.ObjectMeta.Labels, service.Spec.Selector) {
					podStatus := podStatus(pod)
					podLogs := ""
					if podStatus == "CrashLoopBackOff" || podStatus == "Error" {
						podLogs = logs(e, pod)
					}
					deployment.Pods = append(deployment.Pods, &api.Pod{Name: pod.Name, DeploymentName: deployment.Name, Namespace: pod.Namespace, Status: podStatus, StatusDescription: podErrorCause(pod), Logs: podLogs, ImChannelId: service.ObjectMeta.GetAnnotations()[AnnotationOwnerIm], Containers: PodContainers(pod.Spec)})
				}
			}
		}

		belongTo := service.ObjectMeta.GetAnnotations()[AnnotationBelongsTo]
		for _, stack := range stacks {
			if stack.Service.Name == belongTo {
				stack.Deployment.Pods = append(stack.Deployment.Pods, deployment.Pods...)
			}
		}
	}

	return stacks, nil
}

func PodContainers(podSpec v1.PodSpec) (containers []v1.Container) {
	containers = append(containers, podSpec.InitContainers...)
	containers = append(containers, podSpec.Containers...)

	return containers
}

func (e *KubeEnv) FluxState() (*flux.FluxState, error) {
	return flux.State(e.Client.(*kubernetes.Clientset), e.DynamicClient.(*dynamic.DynamicClient))
}

func labelsMatchSelectors(labels map[string]string, selectors map[string]string) bool {
	for k2, v2 := range selectors {
		if v, ok := labels[k2]; ok {
			if v2 != v {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func getOpenServiceCatalogAnnotations(svc v1.Service) *api.Osca {
	return &api.Osca{
		Links: api.Links{
			Docs:    svc.ObjectMeta.GetAnnotations()[AnnotationDocsLink],
			Logs:    svc.ObjectMeta.GetAnnotations()[AnnotationLogsLink],
			Metrics: svc.ObjectMeta.GetAnnotations()[AnnotationMetricsLink],
			Traces:  svc.ObjectMeta.GetAnnotations()[AnnotationTracesLink],
			Issues:  svc.ObjectMeta.GetAnnotations()[AnnotationIssuesLink],
		},
		Owner: svc.ObjectMeta.GetAnnotations()[AnnotationOwnerName],
	}
}

func (e *KubeEnv) FetchCertificate() []byte {
	service, err := e.Client.CoreV1().Services("infrastructure").Get(context.Background(), "sealed-secrets-controller", metav1.GetOptions{})
	if err != nil {
		logrus.Debugf("could not get sealed secret service: %s", err)
		return nil
	}

	cert, err := e.Client.CoreV1().Services("infrastructure").ProxyGet("http", "sealed-secrets-controller", service.Spec.Ports[0].Name, "/v1/cert.pem", nil).DoRaw(context.Background())
	if err != nil {
		logrus.Debugf("could not get cert: %s", err)
		return nil
	}

	return cert
}

var gitRepositoryResource = schema.GroupVersionResource{
	Group:    "source.toolkit.fluxcd.io",
	Version:  "v1",
	Resource: "gitrepositories",
}

var kustomizationResource = schema.GroupVersionResource{
	Group:    "kustomize.toolkit.fluxcd.io",
	Version:  "v1",
	Resource: "kustomizations",
}

var helmReleaseResource = schema.GroupVersionResource{
	Group:    "helm.toolkit.fluxcd.io",
	Version:  "v2beta2",
	Resource: "helmreleases",
}

var tfResource = schema.GroupVersionResource{
	Group:    "infra.contrib.fluxcd.io",
	Version:  "v1alpha2",
	Resource: "terraforms",
}

func (e *KubeEnv) WarningEvents() ([]api.Event, error) {
	integratedServices, err := e.annotatedServices(AnnotationGitRepository)
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
func (e *KubeEnv) annotatedServices(annotation string) ([]v1.Service, error) {
	svc, err := e.Client.CoreV1().Services(e.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []v1.Service
	for _, s := range svc.Items {
		if _, ok := s.ObjectMeta.GetAnnotations()[annotation]; ok {
			services = append(services, s)
		}
	}

	return services, nil
}

func (e *KubeEnv) deploymentForService(service v1.Service, deployments []appsv1.Deployment, statefulsets []appsv1.StatefulSet) (*api.Deployment, error) {
	var deployment *api.Deployment

	for _, d := range deployments {
		if SelectorsMatch(d.Spec.Selector.MatchLabels, service.Spec.Selector) {
			var branch, sha string
			if hash, ok := d.GetAnnotations()[AnnotationGitSha]; ok {
				sha = hash
			}
			if b, ok := d.GetAnnotations()[AnnotationGitBranch]; ok {
				branch = b
			}

			deployment = &api.Deployment{Name: d.Name, Namespace: d.Namespace, Branch: branch, SHA: sha}
		}
	}

	for _, s := range statefulsets {
		if SelectorsMatch(s.Spec.Selector.MatchLabels, service.Spec.Selector) {
			var branch, sha string
			if hash, ok := s.GetAnnotations()[AnnotationGitSha]; ok {
				sha = hash
			}
			if b, ok := s.GetAnnotations()[AnnotationGitBranch]; ok {
				branch = b
			}

			deployment = &api.Deployment{Name: s.Name, Namespace: s.Namespace, Branch: branch, SHA: sha}
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
			if containerStatus.State.Terminated != nil {
				return fmt.Sprint(containerStatus.State.Terminated.Reason)
			}
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
