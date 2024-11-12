package agent

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const terraformCRDName = "terraforms.infra.contrib.fluxcd.io"

func TfController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	return NewDynamicController(
		terraformCRDName,
		kubeEnv.DynamicClient,
		tfResource,
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
