package agent

import (
	"context"

	"github.com/gimlet-io/gimlet/pkg/dashboard/api"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

				allStatefulsets, err := kubeEnv.Client.AppsV1().StatefulSets(kubeEnv.Namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					return err
				}

				createdPod := obj.(*v1.Pod)
				for _, svc := range integratedServices {
					for _, deployment := range allDeployments.Items {
						matchAndSendCreatedEvent(
							deployment.Spec.Selector.MatchLabels,
							deployment.Namespace,
							deployment.Name,
							svc,
							createdPod,
							kubeEnv,
							objectMeta,
							gimletHost,
							agentKey,
						)
					}
					for _, statefulset := range allStatefulsets.Items {
						matchAndSendCreatedEvent(
							statefulset.Spec.Selector.MatchLabels,
							statefulset.Namespace,
							statefulset.Name,
							svc,
							createdPod,
							kubeEnv,
							objectMeta,
							gimletHost,
							agentKey,
						)
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

				allStatefulsets, err := kubeEnv.Client.AppsV1().StatefulSets(kubeEnv.Namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					return err
				}

				if obj == nil {
					return nil
				}

				updatedPod := obj.(*v1.Pod)
				for _, svc := range integratedServices {
					for _, deployment := range allDeployments.Items {
						newFunction(
							deployment.Spec.Selector.MatchLabels,
							deployment.Namespace,
							deployment.Name,
							svc,
							updatedPod,
							kubeEnv,
							objectMeta,
							gimletHost,
							agentKey,
						)
					}
					for _, statefulset := range allStatefulsets.Items {
						newFunction(
							statefulset.Spec.Selector.MatchLabels,
							statefulset.Namespace,
							statefulset.Name,
							svc,
							updatedPod,
							kubeEnv,
							objectMeta,
							gimletHost,
							agentKey,
						)
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

func newFunction(matchLabels map[string]string, namespace string, name string, svc v1.Service, updatedPod *v1.Pod, kubeEnv *KubeEnv, objectMeta metav1.ObjectMeta, gimletHost string, agentKey string) {
	if SelectorsMatch(matchLabels, svc.Spec.Selector) {
		if HasLabels(matchLabels, updatedPod.GetObjectMeta().GetLabels()) &&
			updatedPod.Namespace == namespace {
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

				Status:      podStatus,
				Deployment:  namespace + "/" + name,
				ErrorCause:  podErrorCause(*updatedPod),
				Logs:        podLogs,
				ImChannelId: svc.GetAnnotations()[AnnotationOwnerIm],
			}
			sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
		}
	}
}

func matchAndSendCreatedEvent(matchLabels map[string]string, namespace string, name string, svc v1.Service, createdPod *v1.Pod, kubeEnv *KubeEnv, objectMeta metav1.ObjectMeta, gimletHost string, agentKey string) {
	if SelectorsMatch(matchLabels, svc.Spec.Selector) {
		if HasLabels(matchLabels, createdPod.GetObjectMeta().GetLabels()) &&
			createdPod.Namespace == namespace {
			update := &api.StackUpdate{
				Event:   EventPodCreated,
				Env:     kubeEnv.Name,
				Repo:    svc.GetAnnotations()[AnnotationGitRepository],
				Subject: objectMeta.Namespace + "/" + objectMeta.Name,
				Svc:     svc.Namespace + "/" + svc.Name,

				Status:      string(createdPod.Status.Phase),
				Deployment:  namespace + "/" + name,
				ImChannelId: svc.GetAnnotations()[AnnotationOwnerIm],
			}
			sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
		}
	}
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
