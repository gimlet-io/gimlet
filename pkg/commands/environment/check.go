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
	v1 "k8s.io/api/core/v1"
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
	if err := waitForDeployment(clientSet, "flux-system", "helm-controller", spinner); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "kustomize-controller", spinner); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "notification-controller", spinner); err != nil {
		return spinner.Fail(err)
	}
	if err := waitForDeployment(clientSet, "flux-system", "source-controller", spinner); err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	envName := c.String("env")
	spinner = NewSpinner("Setting up git connection")
	err = waitForResources(client, gitRepositoryResource, envName, spinner)
	if err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Deploying infrastructure components")
	err = waitForResources(client, kustomizationResource, envName, spinner)
	if err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	spinner = NewSpinner("Waiting for Gimlet Agent")
	if err := waitForDeployment(clientSet, "infrastructure", "gimlet-agent", spinner); err != nil {
		return spinner.Fail(err)
	}
	spinner.Success()

	NewSpinner("Done!").Success()

	return nil
}

func waitForDeployment(clientSet *kubernetes.Clientset, namespace, deploymentName string, spinner *Spinner) error {
	for {
		deployment, err := clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, meta_v1.GetOptions{})
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

		pod, err := getDeploymentPod(clientSet, deployment.Spec.Selector.MatchLabels, namespace)
		if err != nil {
			return err
		}

		spinner.Infof("Pod %s/%s: %s", namespace, pod.Name, podStatus(*pod))

		if deployment != nil &&
			deployment.Spec.Replicas != nil &&
			*deployment.Spec.Replicas == deployment.Status.ReadyReplicas {
			return nil
		}
	}
}

func getDeploymentPod(clientSet *kubernetes.Clientset, matchLabels map[string]string, namespace string) (*v1.Pod, error) {
	allPods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range allPods.Items {
		if hasLabels(matchLabels, pod.GetObjectMeta().GetLabels()) &&
			pod.Namespace == namespace {

			return &pod, nil
		}
	}
	return nil, nil
}

func hasLabels(selector map[string]string, labels map[string]string) bool {
	for selectorLabel, selectorValue := range selector {
		hasLabel := false
		for label, value := range labels {
			if label == selectorLabel && value == selectorValue {
				hasLabel = true
			}
		}
		if !hasLabel {
			return false
		}
	}

	return true
}

func podStatus(pod v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating" //https://github.com/kubernetes/kubernetes/issues/61376#issuecomment-374437926
	}

	if v1.PodPending == pod.Status.Phase ||
		v1.PodRunning == pod.Status.Phase {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				return fmt.Sprint(containerStatus.State.Waiting.Reason)
			}
		}
	}

	return fmt.Sprint(pod.Status.Phase)
}

func waitForResources(client dynamic.Interface, gvr schema.GroupVersionResource, envName string, spinner *Spinner) error {
	resources, err := environmentRelatedResources(client, gvr, envName)
	if err != nil {
		return err
	}

	for _, r := range resources {
		err := waitForReadyResource(client, gvr, r.GetName(), spinner)
		if err != nil {
			return err
		}
	}

	return nil
}

func environmentRelatedResources(client dynamic.Interface, gvr schema.GroupVersionResource, envName string) ([]unstructured.Unstructured, error) {
	var relatedResources []unstructured.Unstructured
	resources, err := client.Resource(gvr).Namespace("flux-system").List(context.TODO(), meta_v1.ListOptions{})
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

func waitForReadyResource(client dynamic.Interface, gvr schema.GroupVersionResource, resourceName string, spinner *Spinner) error {
	for {
		resource, err := client.Resource(gvr).Namespace("flux-system").Get(context.TODO(), resourceName, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		ready, reason := getStatus(resource)

		spinner.Infof("Resource %s/%s: %s", gvr.Resource, resourceName, reason)

		if ready {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func getStatus(resource *unstructured.Unstructured) (bool, string) {
	ready := false
	statusMap := resource.Object["status"].(map[string]interface{})
	conditions, _ := statusMap["conditions"].([]interface{})
	reason, status := reasonAndStatus(conditions)
	ready, _ = strconv.ParseBool(status)

	return ready, reason
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
