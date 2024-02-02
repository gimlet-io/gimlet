package agent

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EventKustomizationCreated = "kustomizationCreated"
const EventKustomizationUpdated = "kustomizationUpdated"
const EventKustomizationDeleted = "kustomizationDeleted"

const kustomizationCRDName = "kustomizations.kustomize.toolkit.fluxcd.io"

func KustomizationController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	return NewDynamicController(
		kustomizationCRDName,
		kubeEnv.DynamicClient,
		kustomizationResource,
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			switch informerEvent.eventType {
			case "create":
				fallthrough
			case "update":
				fallthrough
			case "delete":
				SendFluxState(kubeEnv, gimletHost, agentKey)
				SendFluxStatev2(kubeEnv, gimletHost, agentKey)
			}
			return nil
		})
}
