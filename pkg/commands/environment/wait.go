package environment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubectl/pkg/util/interrupt"
)

// RunWait runs the waiting logic
func waitFor(resourceType string) error {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), 60000000000)
	defer cancel()

	f := createFactory()
	namespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	visitCount := 0
	visitFunc := func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		visitCount++
		finalObject, success, err := ConditionalWait{
			conditionName:   "established",
			conditionStatus: "true",
			errOut:          os.Stdout,
		}.IsConditionMet(ctx, info)
		if success {
			printer := &printers.NamePrinter{
				Operation: "condition met",
			}
			printer.PrintObj(finalObject, os.Stdout)
			return nil
		}
		if err == nil {
			return fmt.Errorf("%v unsatisified for unknown reason", finalObject)
		}
		return err
	}

	visitor := f.NewBuilder().
		NamespaceParam(namespace).DefaultNamespace().
		AllNamespaces(false).
		Unstructured().
		FilenameParam(enforceNamespace, &resource.FilenameOptions{}).
		ResourceTypeOrNameArgs(false, resourceType).
		LabelSelectorParam("").
		FieldSelectorParam("").
		Latest().
		ContinueOnError().
		Flatten().
		Do()

	err = visitor.Visit(visitFunc)
	if err != nil {
		return err
	}
	if visitCount == 0 {
		return errors.New("no matching resources found")
	}
	return err
}

// ConditionalWait hold information to check an API status condition
type ConditionalWait struct {
	conditionName   string
	conditionStatus string
	// errOut is written to if an error occurs
	errOut io.Writer
}

// IsConditionMet is a conditionfunc for waiting on an API condition to be met
func (w ConditionalWait) IsConditionMet(ctx context.Context, info *resource.Info) (runtime.Object, bool, error) {
	return getObjAndCheckCondition(ctx, info, w.isConditionMet, w.checkCondition)
}

func (w ConditionalWait) checkCondition(obj *unstructured.Unstructured) (bool, error) {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	for _, conditionUncast := range conditions {
		condition := conditionUncast.(map[string]interface{})
		name, found, err := unstructured.NestedString(condition, "type")
		if !found || err != nil || !strings.EqualFold(name, w.conditionName) {
			continue
		}
		status, found, err := unstructured.NestedString(condition, "status")
		if !found || err != nil {
			continue
		}
		generation, found, _ := unstructured.NestedInt64(obj.Object, "metadata", "generation")
		if found {
			observedGeneration, found := getObservedGeneration(obj, condition)
			if found && observedGeneration < generation {
				return false, nil
			}
		}
		return strings.EqualFold(status, w.conditionStatus), nil
	}

	return false, nil
}

func (w ConditionalWait) isConditionMet(event watch.Event) (bool, error) {
	if event.Type == watch.Error {
		// keep waiting in the event we see an error - we expect the watch to be closed by
		// the server
		err := apierrors.FromObject(event.Object)
		fmt.Fprintf(w.errOut, "error: An error occurred while waiting for the condition to be satisfied: %v", err)
		return false, nil
	}
	if event.Type == watch.Deleted {
		// this will chain back out, result in another get and an return false back up the chain
		return false, nil
	}
	obj := event.Object.(*unstructured.Unstructured)
	return w.checkCondition(obj)
}

func getObservedGeneration(obj *unstructured.Unstructured, condition map[string]interface{}) (int64, bool) {
	conditionObservedGeneration, found, _ := unstructured.NestedInt64(condition, "observedGeneration")
	if found {
		return conditionObservedGeneration, true
	}
	statusObservedGeneration, found, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	return statusObservedGeneration, found
}

type isCondMetFunc func(event watch.Event) (bool, error)
type checkCondFunc func(obj *unstructured.Unstructured) (bool, error)

// getObjAndCheckCondition will make a List query to the API server to get the object and check if the condition is met using check function.
// If the condition is not met, it will make a Watch query to the server and pass in the condMet function
func getObjAndCheckCondition(ctx context.Context, info *resource.Info, condMet isCondMetFunc, check checkCondFunc) (runtime.Object, bool, error) {
	if len(info.Name) == 0 {
		return info.Object, false, fmt.Errorf("resource name must be provided")
	}

	f := createFactory()
	client, err := f.DynamicClient()
	if err != nil {
		return nil, false, err
	}

	endTime := time.Now().Add(60000000000)
	timeout := time.Until(endTime)
	errWaitTimeoutWithName := extendErrWaitTimeout(wait.ErrWaitTimeout, info)
	if timeout < 0 {
		// we're out of time
		return info.Object, false, errWaitTimeoutWithName
	}

	mapping := info.ResourceMapping() // used to pass back meaningful errors if object disappears
	fieldSelector := fields.OneTermEqualSelector("metadata.name", info.Name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return client.Resource(info.Mapping.Resource).Namespace(info.Namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return client.Resource(info.Mapping.Resource).Namespace(info.Namespace).Watch(context.TODO(), options)
		},
	}

	// this function is used to refresh the cache to prevent timeout waits on resources that have disappeared
	preconditionFunc := func(store cache.Store) (bool, error) {
		_, exists, err := store.Get(&metav1.ObjectMeta{Namespace: info.Namespace, Name: info.Name})
		if err != nil {
			return true, err
		}
		if !exists {
			return true, apierrors.NewNotFound(mapping.Resource.GroupResource(), info.Name)
		}

		return false, nil
	}

	intrCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var result runtime.Object
	intr := interrupt.New(nil, cancel)
	err = intr.Run(func() error {
		ev, err := watchtools.UntilWithSync(intrCtx, lw, &unstructured.Unstructured{}, preconditionFunc, watchtools.ConditionFunc(condMet))
		if ev != nil {
			result = ev.Object
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return errWaitTimeoutWithName
		}
		return err
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return result, false, errWaitTimeoutWithName
		}
		return result, false, err
	}

	return result, true, nil
}

func extendErrWaitTimeout(err error, info *resource.Info) error {
	return fmt.Errorf("%s on %s/%s", err.Error(), info.Mapping.Resource.Resource, info.Name)
}
