package agent

import (
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
				logrus.Info("gitRepository created: " + objectMeta.Name)
				gitRepositories, err := kubeEnv.GitRepositories()
				if err != nil {
					return err
				}

				for _, g := range gitRepositories {
					logrus.Info(g)
				}
			case "update":
				logrus.Info("gitRepository updated: " + objectMeta.Name)
				gitRepositories, err := kubeEnv.GitRepositories()
				if err != nil {
					return err
				}

				for _, g := range gitRepositories {
					logrus.Info(g)
				}
			case "delete":
				logrus.Info("gitRepository deleted: " + objectMeta.Name)
				gitRepositories, err := kubeEnv.GitRepositories()
				if err != nil {
					return err
				}

				for _, g := range gitRepositories {
					logrus.Info(g)
				}
			}
			return nil
		})
}
