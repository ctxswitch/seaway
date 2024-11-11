package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubectl/pkg/util/interrupt"
)

// WaitOptions contains the options for used to wait on a condition.
type WaitOptions struct {
	timeout time.Duration
	client  dynamic.ResourceInterface
}

// WaitCondition defines the conditions that needs to be satisfied.
type WaitCondition struct {
	// The name of the resource to wait for.  Used as the field selector for the list/watch.
	name string
	// The condition (or .status.type) to wait for.
	condition string
	// The kind of the object to wait for.
	kind string
}

// NewWaitOptions creates a new WaitOptions with the provided options.
func NewWaitCondition(name, kind, condition string) *WaitCondition {
	return &WaitCondition{name: name, condition: condition, kind: kind}
}

// WaitForCondition creates a listwatch and blocks until the condition has
// been met or a timeout occurs.
func (w *WaitCondition) WaitForCondition(ctx context.Context, opts WaitOptions) error {
	client := opts.client

	options := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", w.name).String(),
	}

	lw := &cache.ListWatch{
		ListFunc: func(metav1.ListOptions) (runtime.Object, error) {
			return client.List(ctx, options)
		},
		WatchFunc: func(metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, options)
		},
	}

	preconditionFunc := func(store cache.Store) (bool, error) {
		return false, nil
	}

	intrCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	intr := interrupt.New(nil, cancel)
	err := intr.Run(func() error {
		_, err := watchtools.UntilWithSync(intrCtx, lw, &unstructured.Unstructured{}, preconditionFunc, watchtools.ConditionFunc(w.isConditionMet))
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timed out waiting for condition %s", w.condition)
		}
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// isConditionMet checks a watch event and passes the object on to
// the check function.
func (w *WaitCondition) isConditionMet(event watch.Event) (bool, error) {
	if event.Type == watch.Error {
		err := apierr.FromObject(event.Object)
		fmt.Fprintf(os.Stderr, "error: waiting for condition to be satisfied %v", err)
		return false, nil
	}
	if event.Type == watch.Deleted {
		return false, nil
	}

	obj, _ := event.Object.(*unstructured.Unstructured)
	return w.checkCondition(obj)
}

// checkCondition parses the condition and routes it to the appropriate
// check function to handle a specific type of check.
func (w *WaitCondition) checkCondition(obj *unstructured.Unstructured) (bool, error) {
	condition := strings.ToLower(w.condition)
	switch {
	case condition == "ready" && !strings.EqualFold(w.kind, "statefulset"):
		return w.checkStatusCondition(obj, "available")
	case condition == "ready" && strings.EqualFold(w.kind, "statefulset"):
		return w.checkStatusReplicas(obj)
	case strings.HasPrefix(condition, "condition="):
		return w.checkStatusCondition(obj, strings.TrimPrefix(condition, "condition="))
	}
	return false, fmt.Errorf("unknown condition %s", condition)
}

// checkStatusCondition checks the condition list in the status block ensuring
// that a condition type is equal to true.
func (w *WaitCondition) checkStatusCondition(obj *unstructured.Unstructured, conditionString string) (bool, error) {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	for _, raw := range conditions {
		condition, ok := raw.(map[string]interface{})
		if !ok {
			return false, nil
		}

		name, found, err := unstructured.NestedString(condition, "type")
		if !found || err != nil || !strings.EqualFold(name, conditionString) {
			continue
		}

		status, found, err := unstructured.NestedString(condition, "status")
		if !found || err != nil {
			continue
		}

		return strings.EqualFold(status, "true"), nil
	}

	return false, nil
}

// checkStatusReplicas is a status check for resources that do not have an
// explicit condition that reports that the resource is ready.  It checks that
// the number of updated and available replicas is equal to the number of
// desired replicas.
func (w *WaitCondition) checkStatusReplicas(obj *unstructured.Unstructured) (bool, error) {
	replicas, found, err := unstructured.NestedFieldNoCopy(obj.Object, "status", "replicas")
	if !found || err != nil {
		return false, nil
	}

	availableReplicas, found, err := unstructured.NestedFieldNoCopy(obj.Object, "status", "availableReplicas")
	if !found || err != nil {
		return false, nil
	}

	updatedReplicas, found, err := unstructured.NestedFieldNoCopy(obj.Object, "status", "updatedReplicas")
	if !found || err != nil {
		return false, nil
	}

	isUpdated := updatedReplicas == replicas
	isAvailable := availableReplicas == replicas

	return isUpdated && isAvailable, nil
}
