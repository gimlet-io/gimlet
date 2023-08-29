package agent

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

const EventPodCreated = "podCreated"
const EventPodUpdated = "podUpdated"
const EventPodDeleted = "podDeleted"

func PodController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	podListWatcher := cache.NewListWatchFromClient(kubeEnv.Client.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
	podController := NewController(
		"pod",
		podListWatcher,
		&v1.Pod{},
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			switch informerEvent.eventType {
			case "create":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				allDeployments, err := kubeEnv.Client.AppsV1().Deployments(kubeEnv.Namespace).List(context.TODO(), meta_v1.ListOptions{})
				if err != nil {
					return err
				}

				createdPod := obj.(*v1.Pod)
				for _, svc := range integratedServices {
					for _, deployment := range allDeployments.Items {
						if SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
							if HasLabels(deployment.Spec.Selector.MatchLabels, createdPod.GetObjectMeta().GetLabels()) &&
								createdPod.Namespace == deployment.Namespace {
								update := &api.StackUpdate{
									Event:   EventPodCreated,
									Env:     kubeEnv.Name,
									Repo:    svc.GetAnnotations()[AnnotationGitRepository],
									Subject: objectMeta.Namespace + "/" + objectMeta.Name,
									Svc:     svc.Namespace + "/" + svc.Name,

									Status:     string(createdPod.Status.Phase),
									Deployment: deployment.Namespace + "/" + deployment.Name,
								}
								sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
							}
						}
					}
				}
			case "update":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				allDeployments, err := kubeEnv.Client.AppsV1().Deployments(kubeEnv.Namespace).List(context.TODO(), meta_v1.ListOptions{})
				if err != nil {
					return err
				}

				if obj == nil {
					return nil
				}

				updatedPod := obj.(*v1.Pod)
				for _, svc := range integratedServices {
					for _, deployment := range allDeployments.Items {
						if SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
							if HasLabels(deployment.Spec.Selector.MatchLabels, updatedPod.GetObjectMeta().GetLabels()) &&
								updatedPod.Namespace == deployment.Namespace {
								podStatus := podStatus(*updatedPod)
								podLogs := ""
								if "CrashLoopBackOff" == podStatus {
									podLogs = logs(kubeEnv, *updatedPod)
								}

								update := &api.StackUpdate{
									Event:   EventPodUpdated,
									Env:     kubeEnv.Name,
									Repo:    svc.GetAnnotations()[AnnotationGitRepository],
									Subject: objectMeta.Namespace + "/" + objectMeta.Name,
									Svc:     svc.Namespace + "/" + svc.Name,

									Status:     podStatus,
									Deployment: deployment.Namespace + "/" + deployment.Name,
									ErrorCause: podErrorCause(*updatedPod),
									Logs:       podLogs,
								}
								sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
							}
						}
					}
				}
			case "delete":
				update := &api.StackUpdate{
					Event:   EventPodDeleted,
					Env:     kubeEnv.Name,
					Subject: informerEvent.key,
				}
				sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
			}
			return nil
		})
	return podController
}

// hasLabels determines if all the selectors are present as labels
func HasLabels(selector map[string]string, labels map[string]string) bool {
	for selectorLabel, selectorValue := range selector {
		hasLabel := false
		for label, value := range labels {
			if label == selectorLabel && value == selectorValue {
				hasLabel = true
			}
		}
		if !hasLabel {
			return false
		}
	}

	return true
}
