package agent

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EventHelmReleaseCreated = "helmReleaseCreated"
const EventHelmReleaseUpdated = "helmReleaseUpdated"
const EventHelmReleaseDeleted = "helmReleaseDeleted"

const helmReleaseCRDName = "helmreleases.helm.toolkit.fluxcd.io"

func HelmReleaseController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	return NewDynamicController(
		helmReleaseCRDName,
		kubeEnv.DynamicClient,
		helmReleaseResource,
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
