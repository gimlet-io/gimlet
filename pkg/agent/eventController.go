package agent

import (
	"context"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func EventController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	eventListWatcher := cache.NewListWatchFromClient(kubeEnv.Client.CoreV1().RESTClient(), "events", v1.NamespaceAll, fields.Everything())
	eventController := NewController(
		"event",
		eventListWatcher,
		&v1.Event{},
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			integratedServices, err := kubeEnv.annotatedServices("")
			if err != nil {
				return err
			}

			allDeployments, err := kubeEnv.Client.AppsV1().Deployments("").List(context.TODO(), meta_v1.ListOptions{})
			if err != nil {
				return err
			}

			allPods, err := kubeEnv.Client.CoreV1().Pods("").List(context.TODO(), meta_v1.ListOptions{})
			if err != nil {
				return err
			}

			events, err := kubeEnv.Client.CoreV1().Events("").List(context.TODO(), meta_v1.ListOptions{})
			if err != nil {
				return err
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

			sendEvents(gimletHost, agentKey, typeWarningEvents)
			return nil
		})
	return eventController
}
