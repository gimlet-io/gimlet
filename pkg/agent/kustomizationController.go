package agent

import (
	"github.com/sirupsen/logrus"
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
				logrus.Info("kustomization created: " + objectMeta.Name)
				kustomizations, err := kubeEnv.Kustomizations()
				if err != nil {
					return err
				}

				for _, k := range kustomizations {
					logrus.Info(k)
				}
			case "update":
				logrus.Info("kustomization updated: " + objectMeta.Name)
				kustomizations, err := kubeEnv.Kustomizations()
				if err != nil {
					return err
				}

				for _, k := range kustomizations {
					logrus.Info(k)
				}
			case "delete":
				logrus.Info("kustomization deleted: " + objectMeta.Name)
				kustomizations, err := kubeEnv.Kustomizations()
				if err != nil {
					return err
				}

				for _, k := range kustomizations {
					logrus.Info(k)
				}
			}
			return nil
		})
}
