package environment

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var environmentCheckCmd = cli.Command{
	Name:      "check",
	Usage:     "Checks if Flux installed and ready on the cluster.",
	UsageText: `gimlet environment check`,
	Action:    check,
}

func check(c *cli.Context) error {
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)
	go func() {
		<-ctrlC
		os.Exit(0)
	}()

	f := createFactory()
	clientSet, err := f.KubernetesClientSet()
	if err != nil {
		return err
	}
	client, err := f.DynamicClient()
	if err != nil {
		return err
	}

	spinner := NewSpinner("Installing Flux")
	if err := waitForDeployment(clientSet, "flux-system", "helm-controller"); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "kustomize-controller"); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "notification-controller"); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "source-controller"); err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Setting up git connection")
	err = waitForResources(client, schema.GroupVersionResource{
		Group:    "source.toolkit.fluxcd.io",
		Version:  "v1beta1",
		Resource: "gitrepositories",
	}, 2)
	if err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Deploying infrastructure components")
	err = waitForResources(client, schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1beta1",
		Resource: "kustomizations",
	}, 3)
	if err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Waiting for Gimlet Agent")
	if err := waitForDeployment(clientSet, "infrastructure", "gimlet-agent"); err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	NewSpinner("Done!").Success()

	return nil
}

func waitForDeployment(clientSet *kubernetes.Clientset, namespace, deploymentName string) error {
	for {
		deploy, err := clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, meta_v1.GetOptions{})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				time.Sleep(3 * time.Second)
				continue
			}
			return err
		}

		if deploy != nil &&
			deploy.Spec.Replicas != nil &&
			*deploy.Spec.Replicas == deploy.Status.ReadyReplicas {
			return nil
		}
	}
}

func waitForResources(client dynamic.Interface, gvr schema.GroupVersionResource, resourcesCount int) error {
	for {
		ready := true
		resources, err := client.Resource(gvr).Namespace("").List(context.TODO(), meta_v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, resource := range resources.Items {
			status := resource.Object["status"].(map[string]interface{})

			conditions, ok := status["conditions"].([]interface{})
			if !ok {
				continue
			}

			for _, condition := range conditions {
				conditionMap := condition.(map[string]interface{})
				conditionStatus := conditionMap["status"].(string)

				if conditionStatus == "False" {
					ready = false
				}
			}
		}

		if ready && len(resources.Items) == resourcesCount {
			return nil
		}
	}
}
