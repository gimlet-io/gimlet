package agent

import (
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	v1 "k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

const EventIngressCreated = "ingressCreated"
const EventIngressUpdated = "ingressUpdated"
const EventIngressDeleted = "ingressDeleted"

func IngressController(kubeEnv *KubeEnv, gimletHost string, agentKey string) *Controller {
	ingressListWatcher := cache.NewListWatchFromClient(kubeEnv.Client.ExtensionsV1beta1().RESTClient(), "ingresses", v1.NamespaceAll, fields.Everything())
	ingressController := NewController(
		"ingress",
		ingressListWatcher,
		&ext_v1beta1.Ingress{},
		func(informerEvent Event, objectMeta meta_v1.ObjectMeta, obj interface{}) error {
			switch informerEvent.eventType {
			case "create":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				createdIngress := obj.(*ext_v1beta1.Ingress)
				for _, svc := range integratedServices {
					for _, rule := range createdIngress.Spec.Rules {
						for _, path := range rule.HTTP.Paths {
							if path.Backend.ServiceName == svc.Name &&
								createdIngress.Namespace == svc.Namespace {
								update := &api.StackUpdate{
									Event:   EventIngressCreated,
									Env:     kubeEnv.Name,
									Repo:    svc.GetAnnotations()[AnnotationGitRepository],
									Subject: objectMeta.Namespace + "/" + objectMeta.Name,
									Svc:     svc.Namespace + "/" + svc.Name,

									URL: rule.Host,
								}
								sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
							}
						}
					}
				}
			case "update":
				integratedServices, err := kubeEnv.annotatedServices("")
				if err != nil {
					return err
				}

				ingress := obj.(*ext_v1beta1.Ingress)
				for _, svc := range integratedServices {
					for _, rule := range ingress.Spec.Rules {
						for _, path := range rule.HTTP.Paths {
							if path.Backend.ServiceName == svc.Name {
								update := &api.StackUpdate{
									Event:   EventIngressUpdated,
									Env:     kubeEnv.Name,
									Repo:    svc.GetAnnotations()[AnnotationGitRepository],
									Subject: objectMeta.Namespace + "/" + objectMeta.Name,
									Svc:     svc.Namespace + "/" + svc.Name,

									URL: rule.Host,
								}
								sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
							}
						}
					}
				}
			case "delete":
				update := &api.StackUpdate{
					Event:   EventIngressDeleted,
					Env:     kubeEnv.Name,
					Subject: informerEvent.key,
				}
				sendUpdate(gimletHost, agentKey, kubeEnv.Name, update)
			}
			return nil
		})
	return ingressController
}
