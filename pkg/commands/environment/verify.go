package environment

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var environmentVerifyCmd = cli.Command{
	Name:      "verify",
	Usage:     "Check Flux's custom resources on the cluster to verify the gitops automation.",
	UsageText: `gimlet environment verify`,
	Action:    verify,
}

func verify(c *cli.Context) error {
	gitRepositoriesList, err := getResourceList(schema.GroupVersionResource{
		Group:    "source.toolkit.fluxcd.io",
		Version:  "v1beta1",
		Resource: "gitrepositories",
	})
	if err != nil {
		return err
	}
	fmt.Println(gitRepositoriesList)

	kustomizationList, err := getResourceList(schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1beta1",
		Resource: "kustomizations",
	})
	if err != nil {
		return err
	}
	fmt.Println(kustomizationList)
	return nil
}

func getResourceList(gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	f := createFactory()
	client, err := f.DynamicClient()
	if err != nil {
		return nil, err
	}

	return client.Resource(gvr).Namespace("").List(context.TODO(), meta_v1.ListOptions{})
}
