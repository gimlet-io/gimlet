package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EventGitRepositoryCreated = "gitRepositoryCreated"
const EventGitRepositoryUpdated = "gitRepositoryUpdated"
const EventGitRepositoryDeleted = "gitRepositoryDeleted"

const gitRepositoryCRDName = "gitrepositories.source.toolkit.fluxcd.io"

func GitRepositoryController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	return NewDynamicController(
		gitRepositoryCRDName,
		kubeEnv.DynamicClient,
		gitRepositoryResource,
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			switch informerEvent.eventType {
			case "create":
				fallthrough
			case "update":
				fallthrough
			case "delete":
				SendFluxState(kubeEnv, gimletHost, agentKey)
			}
			return nil
		})
}

func SendFluxState(kubeEnv *KubeEnv, gimletHost string, agentKey string) {
	fluxState, err := kubeEnv.FluxState()
	if err != nil {
		logrus.Errorf("could not get flux state: %s", err)
		return
	}

	fluxStateString, err := json.Marshal(fluxState)
	if err != nil {
		logrus.Errorf("could not serialize flux state: %v", err)
		return
	}

	params := url.Values{}
	params.Add("name", kubeEnv.Name)
	reqUrl := fmt.Sprintf("%s/agent/fluxState?%s", gimletHost, params.Encode())
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(fluxStateString))
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
