// +build fixtures

package agent

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"testing"
	"time"
)

func Test_CreatePod(t *testing.T) {
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	k8sConfig, err := k8sConfig()
	clientset, err := kubernetes.NewForConfig(k8sConfig)

	_, err = clientset.CoreV1().Pods("default").Create(context.TODO(), p, metav1.CreateOptions{})
	if err != nil {
		log.Error(err)
	}
	time.Sleep(3 * time.Second)

	err = clientset.CoreV1().Pods("default").Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Error(err)
	}
}

func k8sConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return config, err
}
