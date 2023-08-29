package environment

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/urfave/cli/v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var gitRepositoryResource = schema.GroupVersionResource{
	Group:    "source.toolkit.fluxcd.io",
	Version:  "v1beta1",
	Resource: "gitrepositories",
}

var kustomizationResource = schema.GroupVersionResource{
	Group:    "kustomize.toolkit.fluxcd.io",
	Version:  "v1beta1",
	Resource: "kustomizations",
}

var environmentCheckCmd = cli.Command{
	Name:      "check",
	Usage:     "Checks if Flux installed and ready on the cluster.",
	UsageText: `gimlet environment check`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to check connection with the cluster",
			Required: true,
		},
	},
	Action: check,
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

	envName := c.String("env")
	spinner = NewSpinner("Setting up git connection")
	err = waitForResources(client, spinner, gitRepositoryResource, envName)
	if err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Deploying infrastructure components")
	err = waitForResources(client, spinner, kustomizationResource, envName)
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

func waitForResources(client dynamic.Interface, spinner *Spinner, resource schema.GroupVersionResource, envName string) error {
	resources, err := relatedResources(client, resource, envName)
	if err != nil {
		return err
	}

	for _, r := range resources {
		err := waitForReadyResource(client, spinner, resource, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForReadyResource(client dynamic.Interface, spinner *Spinner, resource schema.GroupVersionResource, r unstructured.Unstructured) error {
	for {
		updatedResource, err := getUpdatedResource(client, resource, r)
		if err != nil {
			return err
		}

		ready, reason, err := getResourceStatus(updatedResource)
		if err != nil {
			return err
		}

		spinner.Infof("Resource %s/%s: %s", resource.Resource, r.GetName(), reason)

		if ready {
			break
		}
	}

	return nil
}

func getUpdatedResource(client dynamic.Interface, resource schema.GroupVersionResource, r unstructured.Unstructured) (unstructured.Unstructured, error) {
	updatedResources, err := client.Resource(resource).Namespace("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	for _, updatedResource := range updatedResources.Items {
		if updatedResource.GetName() == r.GetName() {
			return updatedResource, nil
		}
	}

	return unstructured.Unstructured{}, fmt.Errorf("resource not found")
}

func getResourceStatus(resource unstructured.Unstructured) (bool, string, error) {
	statusMap := resource.Object["status"].(map[string]interface{})
	conditions, ok := statusMap["conditions"].([]interface{})
	if !ok {
		return false, "", fmt.Errorf("status conditions not found")
	}

	reason, status := reasonAndStatus(conditions)
	ready, _ := strconv.ParseBool(status)
	if ready {
		return true, reason, nil
	}

	return false, reason, nil
}

func relatedResources(client dynamic.Interface, resource schema.GroupVersionResource, envName string) ([]unstructured.Unstructured, error) {
	var relatedResources []unstructured.Unstructured
	resources, err := client.Resource(resource).Namespace("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return relatedResources, err
	}

	for _, r := range resources.Items {
		if strings.Contains(r.GetName(), envName) {
			relatedResources = append(relatedResources, r)
		}
	}

	return relatedResources, nil
}

func reasonAndStatus(conditions []interface{}) (string, string) {
	if c := findStatusCondition(conditions, meta.ReadyCondition); c != nil {
		return c["reason"].(string), c["status"].(string)
	}
	return string(meta_v1.ConditionFalse), "waiting to be reconciled"
}

func findStatusCondition(conditions []interface{}, conditionType string) map[string]interface{} {
	for _, c := range conditions {
		cMap := c.(map[string]interface{})
		if cMap["type"] == conditionType {
			return cMap
		}
	}

	return nil
}
