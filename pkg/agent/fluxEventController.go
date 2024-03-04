package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gimlet-io/capacitor/pkg/flux"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func FluxEventController(
	kubeEnv *KubeEnv, gimletHost string, agentKey string,
) *Controller {
	eventListWatcher := cache.NewListWatchFromClient(kubeEnv.Client.CoreV1().RESTClient(), "events", v1.NamespaceAll, fields.Everything())
	eventController := NewController(
		"event",
		eventListWatcher,
		&v1.Event{},
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			if _, ok := obj.(*v1.Event); !ok {
				return nil
			}

			if flux.IgnoreEvent(*obj.(*v1.Event)) {
				return nil
			}

			SendFluxK8sEvents(kubeEnv, gimletHost, agentKey)
			return nil
		})
	return eventController
}

func SendFluxK8sEvents(kubeEnv *KubeEnv, gimletHost string, agentKey string) {
	events, err := flux.Events(kubeEnv.Client.(*kubernetes.Clientset), kubeEnv.DynamicClient.(*dynamic.DynamicClient))
	if err != nil {
		logrus.Errorf("could not get flux events: %s", err)
		return
	}

	fluxEventsString, err := json.Marshal(events)
	if err != nil {
		logrus.Errorf("could not serialize flux state: %v", err)
		return
	}

	params := url.Values{}
	params.Add("name", kubeEnv.Name)
	reqUrl := fmt.Sprintf("%s/agent/fluxEvents?%s", gimletHost, params.Encode())
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(fluxEventsString))
	if err != nil {
		logrus.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("could not send flux state: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("could not send flux state: %d - %v", resp.StatusCode, string(body))
		return
	}
}
