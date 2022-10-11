package agent

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

const EventDeploymentCreated = "deploymentCreated"
const EventDeploymentUpdated = "deploymentUpdated"
const EventDeploymentDeleted = "deploymentDeleted"

func DeploymentController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	deploymentListWatcher := cache.NewListWatchFromClient(kubeEnv.Client.AppsV1().RESTClient(), "deployments", v1.NamespaceAll, fields.Everything())
	deploymentController := NewController(
		"deployment",
		deploymentListWatcher,
		&apps_v1.Deployment{},
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			switch informerEvent.eventType {
			case "create":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				createdDeployment := obj.(*apps_v1.Deployment)
				for _, svc := range integratedServices {
					if SelectorsMatch(createdDeployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						var sha string
						if hash, ok := createdDeployment.GetAnnotations()[AnnotationGitSha]; ok {
							sha = hash
						}

						update := &api.StackUpdate{
							Event:   EventDeploymentCreated,
							Env:     kubeEnv.Name,
							Repo:    svc.GetAnnotations()[AnnotationGitRepository],
							Subject: objectMeta.Namespace + "/" + objectMeta.Name,
							Svc:     svc.Namespace + "/" + svc.Name,

							SHA: sha,
						}
						sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
					}
				}
			case "update":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				updatedDeployment := obj.(*apps_v1.Deployment)
				for _, svc := range integratedServices {
					if SelectorsMatch(updatedDeployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						var sha string
						if hash, ok := updatedDeployment.GetAnnotations()[AnnotationGitSha]; ok {
							sha = hash
						}

						update := &api.StackUpdate{
							Event:   EventDeploymentUpdated,
							Env:     kubeEnv.Name,
							Repo:    svc.GetAnnotations()[AnnotationGitRepository],
							Subject: objectMeta.Namespace + "/" + objectMeta.Name,
							Svc:     svc.Namespace + "/" + svc.Name,

							SHA: sha,
						}
						sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
					}
				}
			case "delete":
				update := &api.StackUpdate{
					Event:   EventDeploymentDeleted,
					Env:     kubeEnv.Name,
					Subject: informerEvent.key,
				}
				sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
			}
			return nil
		})
	return deploymentController
}
