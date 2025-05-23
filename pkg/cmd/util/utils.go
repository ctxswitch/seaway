package util

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	watchtools "k8s.io/client-go/tools/watch"
)

var (
	DefaultBackoff = wait.Backoff{ //nolint:gochecknoglobals
		Steps:    5,
		Duration: 200 * time.Millisecond,
		Factor:   2.0,
	}
)

// Apply creates or updates a kubernetes resource.
func Apply(ctx context.Context, client *kube.KubectlCmd, k kustomize.KustomizerResource) error {
	obj := k.Resource.DeepCopy()
	api := ToAPIString(obj)

	var op kube.OperationResult
	var err error

	expected, err := kube.ConvertToUnstructured(obj)

	opFunc := func() error {
		var ferr error
		current, ferr := kube.ConvertToUnstructured(obj)
		if ferr != nil {
			return ferr
		}

		currentJson, ferr := json.Marshal(current)
		if ferr != nil {
			return ferr
		}
		expectedJson, ferr := json.Marshal(expected)
		if ferr != nil {
			return ferr
		}

		modifiedJson, ferr := jsonpatch.MergeMergePatches(currentJson, expectedJson)
		if ferr != nil {
			return err
		}

		var modified unstructured.Unstructured
		ferr = json.Unmarshal(modifiedJson, &modified)
		if ferr != nil {
			return ferr
		}

		ferr = kube.ConvertFromUnstructured(&modified, obj)
		if ferr != nil {
			return ferr
		}

		return nil
	}

	err = wait.ExponentialBackoffWithContext(ctx, DefaultBackoff, func(context.Context) (bool, error) {
		op, err = client.CreateOrUpdate(ctx, obj, opFunc)
		if err == nil {
			return true, nil
		}

		if apierr.IsNotFound(err) {
			return false, nil
		}

		return false, err
	})
	if err != nil {
		console.Fatal("error applying resource %s: %s", api, err.Error())
		return err
	}

	switch op {
	case kube.OperationResultNone:
		console.Unchanged(api)
	case kube.OperationResultUpdated:
		console.Updated(api)
	case kube.OperationResultCreated:
		console.Created(api)
	}

	return nil
}

// Wait blocks until a resource meets the defined conditions.
func Wait(ctx context.Context, client *kube.KubectlCmd, k kustomize.KustomizerResource, cond v1beta1.ManifestWaitCondition) error {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, cond.Timeout)
	defer cancel()

	obj := k.Resource

	out := ToAPIString(obj)
	console.Waiting(out)

	err := client.WaitForCondition(ctx, obj, cond.For, cond.Timeout)
	if err != nil {
		return err
	}

	return nil
}

// ToAPIString returns the gvk, namespace, and name of the object as a string.
func ToAPIString(obj *unstructured.Unstructured) string {
	api := strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
	var out string
	if obj.GetNamespace() == "" {
		out = fmt.Sprintf("%s/%s", api, obj.GetName())
	} else {
		out = fmt.Sprintf("%s/%s/%s", api, obj.GetNamespace(), obj.GetName())
	}

	return out
}

// GetEnvironment returns a new environment object.
func GetEnvironment(name, namespace string) *v1beta1.Environment {
	env := &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	// TODO: We actually need this with the client at this point because we use
	// the gvk to get the resource interface. Revisit this later and refactor it
	// out.  It's not horrible but it's not great either.
	env.SetGroupVersionKind(v1beta1.SchemeGroupVersion.WithKind("Environment"))

	return env
}

// GetNamespace returns a new namespace object.
func GetNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ns.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Namespace"))

	return ns
}
