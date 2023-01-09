package notifications

import (
	"database/sql"
	"fmt"
	"time"

	gimletdConfig "github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/agent"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/sirupsen/logrus"
)

type PodStateManager struct {
	NotifManager Manager
}

func NewPodStateManager(notifManager Manager) *PodStateManager {
	return &PodStateManager{NotifManager: notifManager}
}

func (p PodStateManager) Start(kubeEnv *agent.KubeEnv) {
	gimletdConfig, err := gimletdConfig.Environ()
	if err != nil {
		logrus.Fatalln("main: invalid configuration")
	}

	store := store.New(gimletdConfig.Database.Driver, gimletdConfig.Database.Config)

	go func() {
		for {
			// annotatedServices, err := kubeEnv.annotatedServices("")
			// if err != nil {
			// 	logrus.Errorf("could not get 1 %v", err)
			// 	return
			// }

			// d, err := kubeEnv.Client.AppsV1().Deployments(kubeEnv.Namespace).List(context.TODO(), metav1.ListOptions{})
			// if err != nil {
			// 	logrus.Errorf("could not get deployments: %s", err)
			// 	return
			// }

			// for _, service := range annotatedServices {
			// 	deployment, err := kubeEnv.deploymentForService(service, d.Items)
			// 	if err != nil {
			// 		logrus.Errorf("could not get deployment for service: %s", err)
			// 		return
			// 	}

			p.trackStates([]api.Pod{}, *store)
			// 	p.trackStates(deployment.Pods, *store)
			// }

			time.Sleep(1 * time.Minute)
		}
	}()
}

func (p PodStateManager) trackStates(pods []api.Pod, store store.Store) {
	for _, pod := range pods {
		deployment := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podFromStore, err := store.Pod(deployment)
		if err == sql.ErrNoRows {
			if podErrorState(pod.Status) {
				//TODO
				// p.NotifManager.Broadcast(msg)
				err = store.SaveOrUpdatePod(&model.Pod{})
			}
			continue
		} else if err != nil {
			logrus.Errorf("couldn't get pod from db: %s", err)
			continue
		}

		if podErrorState(pod.Status) && podFromStore.Status == "" || podErrorState(pod.Status) && pod.Status != podFromStore.Status {
			//TODO
			// p.NotifManager.Broadcast(msg)
			err = store.SaveOrUpdatePod(&model.Pod{})
		}
	}
}

func podErrorState(status string) bool {
	return status != "Running" && status != "Pending" && status != "Terminating" &&
		status != "Succeeded" && status != "Unknown" && status != "ContainerCreating" &&
		status != "PodInitializing"
}
