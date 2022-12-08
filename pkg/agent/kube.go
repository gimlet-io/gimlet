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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const AnnotationGitRepository = "gimlet.io/git-repository"
const AnnotationGitSha = "gimlet.io/git-sha"

type KubeEnv struct {
	Name      string
	Namespace string
	Client    kubernetes.Interface
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

func KubeEvents(kubeEnv *KubeEnv, gimletHost string, agentKey string) {
	integratedServices, err := kubeEnv.annotatedServices("")
	if err != nil {
		log.Errorf("could not get integrated services: %v", err)
		return
	}

	allDeployments, err := kubeEnv.Client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("could not get deployments: %v", err)
		return
	}

	allPods, err := kubeEnv.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("could not get pods: %v", err)
		return
	}

	events, err := kubeEnv.Client.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("could not get events: %v", err)
		return
	}

	var typeWarningEvents []api.Alert
	for _, svc := range integratedServices {
		for _, deployment := range allDeployments.Items {
			if SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
				for _, pod := range allPods.Items {
					if HasLabels(deployment.Spec.Selector.MatchLabels, pod.GetObjectMeta().GetLabels()) &&
						pod.Namespace == deployment.Namespace {
						for _, event := range events.Items {
							if event.Type == "Warning" && (event.InvolvedObject.Name == pod.Name || event.InvolvedObject.Name == deployment.Name) {
								typeWarningEvents = append(typeWarningEvents, api.Alert{
									LastSeen:            event.LastTimestamp.Unix(),
									DeploymentName:      deployment.Name,
									DeploymentNamespace: deployment.Namespace,
									Reason:              event.Reason,
									Object:              event.InvolvedObject.Name,
									Message:             event.Message,
								})
							}
						}
					}
				}
			}
		}
	}

	typeWarningEventsString, err := json.Marshal(typeWarningEvents)
	if err != nil {
		log.Errorf("could not serialize k8s events: %v", err)
		return
	}

	reqUrl := fmt.Sprintf("%s/agent/events", gimletHost)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(typeWarningEventsString))
	if err != nil {
		log.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Errorf("could not send k8s events: %d - %v", resp.StatusCode, string(body))
		return
	}

	log.Debug("events sent")
}

func IrregularPods(kubeEnv *KubeEnv, gimletHost string, agentKey string) {
	integratedServices, err := kubeEnv.annotatedServices("")
	if err != nil {
		log.Errorf("could not get integrated services: %v", err)
		return
	}

	allDeployments, err := kubeEnv.Client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("could not get deployments: %v", err)
		return
	}

	allPods, err := kubeEnv.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("could not get pods: %v", err)
		return
	}

	var irregularPods []api.Alert
	for _, svc := range integratedServices {
		for _, deployment := range allDeployments.Items {
			if SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
				for _, pod := range allPods.Items {
					if HasLabels(deployment.Spec.Selector.MatchLabels, pod.GetObjectMeta().GetLabels()) &&
						pod.Namespace == deployment.Namespace {
						podStatus := podStatus(pod)

						if isError(podStatus) {
							irregularPods = append(irregularPods, api.Alert{
								DeploymentName:      deployment.Name,
								DeploymentNamespace: deployment.Namespace,
								Reason:              podStatus,
								Object:              pod.Name,
								Message:             podErrorCause(pod),
							})
						}
					}
				}
			}
		}
	}

	irregularPodsString, err := json.Marshal(irregularPods)
	if err != nil {
		log.Errorf("could not serialize k8s events: %v", err)
		return
	}

	reqUrl := fmt.Sprintf("%s/agent/irregularPods", gimletHost)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(irregularPodsString))
	if err != nil {
		log.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Errorf("could not send k8s events: %d - %v", resp.StatusCode, string(body))
		return
	}

	log.Debug("irregular pods sent")
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
			if strings.Contains(s.ObjectMeta.GetAnnotations()[AnnotationGitRepository], "ml-search") {
				continue
			}
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
				pods = append(pods, &api.Pod{Name: pod.Name, Namespace: pod.Namespace, Status: podStatus, StatusDescription: podErrorCause(pod), Logs: podLogs})
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

func isError(podStatus string) bool {
	normalStates := []string{"Running", "Pending", "Terminating", "Succeeded", "Unknown", "ContainerCreating", "PodInitializing"}
	for _, normalState := range normalStates {
		if podStatus == normalState {
			return false
		}
	}
	return true
}
