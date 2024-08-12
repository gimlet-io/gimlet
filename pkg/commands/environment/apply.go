package environment

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jonboulle/clockwork"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	oapi "k8s.io/kube-openapi/pkg/util/proto"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/openapi"
)

const (
	// maxPatchRetry is the maximum number of conflicts retry for during a patch operation before returning failure
	maxPatchRetry = 5
	// backOffPeriod is the period to back off when apply patch results in error.
	backOffPeriod = 1 * time.Second
	// how many times we can retry before back off
	triesBeforeBackOff = 1
)

// Factory provides abstractions that allow the Kubectl command to be extended across multiple types
// of resources and different API sets.
func createFactory() cmdutil.Factory {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	return cmdutil.NewFactory(matchVersionKubeConfigFlags)
}

func getObjects(filePath string) ([]*resource.Info, error) {
	f := createFactory()
	validator, err := f.Validator(metav1.FieldValidationStrict)
	if err != nil {
		return nil, err
	}

	namespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	builder := f.NewBuilder()
	r := builder.
		Unstructured().
		Schema(validator).
		ContinueOnError().
		NamespaceParam(namespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &resource.FilenameOptions{
			Filenames: []string{filePath},
		}).
		LabelSelectorParam("").
		Flatten().
		Do()

	if r.Err() != nil {
		return nil, r.Err()
	}

	return r.Infos()
}

func applyObject(info *resource.Info) (string, error) {
	helper := resource.NewHelper(info.Client, info.Mapping).
		DryRun(false).
		WithFieldManager("kubectl-client-side-apply").
		WithFieldValidation("Strict")

	patcher, err := newPatcher(info, helper)
	if err != nil {
		return "", err
	}

	// Get the modified configuration of the object. Embed the result
	// as an annotation in the modified configuration, so that it will appear
	// in the patch sent to the server.
	modified, err := util.GetModifiedConfiguration(info.Object, true, unstructured.UnstructuredJSONScheme)
	if err != nil {
		return "", err
	}

	if err := info.Get(); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}

		// Create the resource if it doesn't exist
		// First, update the annotation used by kubectl apply
		if err := util.CreateApplyAnnotation(info.Object, unstructured.UnstructuredJSONScheme); err != nil {
			return "", err
		}

		// Then create the resource and skip the three-way merge
		obj, err := helper.Create(info.Namespace, true, info.Object)
		if err != nil {
			return "", err
		}
		info.Refresh(obj, true)

		return fmt.Sprintf("%s created", info.Name), nil
	}

	patchBytes, patchedObject, err := patcher.Patch(info.Object, modified, info.Namespace, info.Name)
	if err != nil {
		return "", err
	}

	if string(patchBytes) == "{}" {
		return fmt.Sprintf("%s unchanged", info.Name), nil
	}

	info.Refresh(patchedObject, true)
	return fmt.Sprintf("%s configured", info.Name), nil
}

// All this code copied from
// https://github.com/kubernetes/kubectl/blob/4ceef69fbc451d9bde6f4d5f92d55624b748141d/pkg/cmd/apply/patcher.go
type Patcher struct {
	Mapping *meta.RESTMapping
	Helper  *resource.Helper

	Overwrite bool
	BackOff   clockwork.Clock

	Force       bool
	Cascade     bool
	Timeout     time.Duration
	GracePeriod int

	// If set, forces the patch against a specific resourceVersion
	ResourceVersion *string

	// Number of retries to make if the patch fails with conflict
	Retries int

	OpenapiSchema openapi.Resources
}

func newPatcher(info *resource.Info, helper *resource.Helper) (*Patcher, error) {
	var openapiSchema openapi.Resources

	return &Patcher{
		Mapping:       info.Mapping,
		Helper:        helper,
		Overwrite:     true,
		BackOff:       clockwork.NewRealClock(),
		Force:         false,
		Cascade:       true,
		Timeout:       time.Duration(0),
		GracePeriod:   -1,
		OpenapiSchema: openapiSchema,
		Retries:       0,
	}, nil
}

func (p *Patcher) Patch(current runtime.Object, modified []byte,
	namespace, name string) ([]byte, runtime.Object, error) {
	var getErr error

	patchBytes, patchObject, err := p.patchSimple(current, modified, namespace, name)

	if p.Retries == 0 {
		p.Retries = maxPatchRetry
	}

	for i := 1; i <= p.Retries && apierrors.IsConflict(err); i++ {
		if i > triesBeforeBackOff {
			p.BackOff.Sleep(backOffPeriod)
		}

		current, getErr = p.Helper.Get(namespace, name)
		if getErr != nil {
			return nil, nil, getErr
		}

		patchBytes, patchObject, err = p.patchSimple(current, modified, namespace, name)
	}

	if err != nil && (apierrors.IsConflict(err) || apierrors.IsInvalid(err)) && p.Force {
		patchBytes, patchObject, err = p.deleteAndCreate(current, modified, namespace, name)
	}

	return patchBytes, patchObject, err
}

func (p *Patcher) patchSimple(obj runtime.Object, modified []byte, namespace, name string) ([]byte, runtime.Object, error) {
	// Serialize the current configuration of the object from the server.
	current, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
	if err != nil {
		return nil, nil, err
	}

	// Retrieve the original configuration of the object from the annotation.
	original, err := util.GetOriginalConfiguration(obj)
	if err != nil {
		return nil, nil, err
	}

	var patchType types.PatchType
	var patch []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta
	var schema oapi.Schema
	// createPatchErrFormat := "creating patch with:\noriginal:\n%s\nmodified:\n%s\ncurrent:\n%s\nfor:"

	// Create the versioned struct from the type defined in the restmapping
	// (which is the API version we'll be submitting the patch to)
	versionedObject, err := scheme.Scheme.New(p.Mapping.GroupVersionKind)
	switch {
	case runtime.IsNotRegisteredError(err):
		// fall back to generic JSON merge patch
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			if mergepatch.IsPreconditionFailed(err) {
				return nil, nil, fmt.Errorf("%s", "At least one of apiVersion, kind and name was changed")
			}
			return nil, nil, err
		}
	case err != nil:
		return nil, nil, err
	case err == nil:
		// Compute a three way strategic merge patch to send to server.
		patchType = types.StrategicMergePatchType

		// Try to use openapi first if the openapi spec is available and can successfully calculate the patch.
		// Otherwise, fall back to baked-in types.
		if p.OpenapiSchema != nil {
			if schema = p.OpenapiSchema.LookupResource(p.Mapping.GroupVersionKind); schema != nil {
				lookupPatchMeta = strategicpatch.PatchMetaFromOpenAPI{Schema: schema}
				if openapiPatch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.Overwrite); err != nil {
					fmt.Fprintf(os.Stderr, "warning: error calculating patch from openapi spec: %v\n", err)
				} else {
					patchType = types.StrategicMergePatchType
					patch = openapiPatch
				}
			}
		}

		if patch == nil {
			lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
			if err != nil {
				return nil, nil, err
			}
			patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.Overwrite)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if string(patch) == "{}" {
		return patch, obj, nil
	}

	if p.ResourceVersion != nil {
		patch, err = addResourceVersion(patch, *p.ResourceVersion)
		if err != nil {
			return nil, nil, err
		}
	}

	patchedObj, err := p.Helper.Patch(namespace, name, patchType, patch, nil)
	return patch, patchedObj, err
}

func (p *Patcher) deleteAndCreate(original runtime.Object, modified []byte, namespace, name string) ([]byte, runtime.Object, error) {
	if err := p.delete(namespace, name); err != nil {
		return modified, nil, err
	}
	// TODO: use wait
	if err := wait.PollImmediate(1*time.Second, p.Timeout, func() (bool, error) {
		if _, err := p.Helper.Get(namespace, name); !apierrors.IsNotFound(err) {
			return false, err
		}
		return true, nil
	}); err != nil {
		return modified, nil, err
	}
	versionedObject, _, err := unstructured.UnstructuredJSONScheme.Decode(modified, nil, nil)
	if err != nil {
		return modified, nil, err
	}
	createdObject, err := p.Helper.Create(namespace, true, versionedObject)
	if err != nil {
		// restore the original object if we fail to create the new one
		// but still propagate and advertise error to user
		recreated, recreateErr := p.Helper.Create(namespace, true, original)
		if recreateErr != nil {
			err = fmt.Errorf("an error occurred force-replacing the existing object with the newly provided one:\n\n%v.\n\nAdditionally, an error occurred attempting to restore the original object:\n\n%v", err, recreateErr)
		} else {
			createdObject = recreated
		}
	}
	return modified, createdObject, err
}

func (p *Patcher) delete(namespace, name string) error {
	options := asDeleteOptions(p.Cascade, p.GracePeriod)
	_, err := p.Helper.DeleteWithOptions(namespace, name, &options)
	return err
}

func asDeleteOptions(cascade bool, gracePeriod int) metav1.DeleteOptions {
	options := metav1.DeleteOptions{}
	if gracePeriod >= 0 {
		options = *metav1.NewDeleteOptions(int64(gracePeriod))
	}
	policy := metav1.DeletePropagationForeground
	if !cascade {
		policy = metav1.DeletePropagationOrphan
	}
	options.PropagationPolicy = &policy
	return options
}

func addResourceVersion(patch []byte, rv string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patch, &patchMap)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{Object: patchMap}
	a, err := meta.Accessor(&u)
	if err != nil {
		return nil, err
	}
	a.SetResourceVersion(rv)

	return json.Marshal(patchMap)
}
