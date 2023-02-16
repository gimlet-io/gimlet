package agent

import (
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
			events, err := kubeEnv.WarningEvents("")
			if err != nil {
				return err
			}

			sendEvents(gimletHost, agentKey, events)
			return nil
		})
	return eventController
}
