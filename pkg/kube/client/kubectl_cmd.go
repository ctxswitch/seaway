// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

const (
	DefaultWaitPollInterval = 500 * time.Millisecond
)

// KubectlCmd is the kubernetes client that is used by the seactl tool.
type KubectlCmd struct {
	client *Client
}

func NewKubectlCmd(ns, context string) (*KubectlCmd, error) {
	c, err := NewClient(ns, context)
	if err != nil {
		return nil, err
	}

	return &KubectlCmd{
		client: c,
	}, nil
}

func (c *KubectlCmd) Get(ctx context.Context, obj Object, opts metav1.GetOptions) error {
	u := &unstructured.Unstructured{}
	if err := toUnstructured(obj, u); err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, obj, "get")
	if err != nil {
		return err
	}

	o, err := iface.Get(ctx, u.GetName(), opts)
	if err != nil {
		return err
	}

	if err := fromUnstructured(o, obj); err != nil {
		return err
	}

	return nil
}

func (c *KubectlCmd) Delete(ctx context.Context, obj Object, opts metav1.DeleteOptions) error {
	u := &unstructured.Unstructured{}
	if err := toUnstructured(obj, u); err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, obj, "delete")
	if err != nil {
		return err
	}

	if opts.PropagationPolicy == nil {
		propagationPolicy := metav1.DeletePropagationForeground
		opts = metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}
	}

	return iface.Delete(ctx, u.GetName(), opts)
}

func (c *KubectlCmd) Create(ctx context.Context, obj Object, opts metav1.CreateOptions) error {
	u := &unstructured.Unstructured{}
	if err := toUnstructured(obj, u); err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, obj, "create")
	if err != nil {
		return err
	}

	o, err := iface.Create(ctx, u, opts)
	if err != nil {
		return err
	}

	err = fromUnstructured(o, obj)
	if err != nil {
		return err
	}

	return err
}

func (c *KubectlCmd) Update(ctx context.Context, obj Object, opts metav1.UpdateOptions) error {
	u := &unstructured.Unstructured{}
	err := toUnstructured(obj, u)
	if err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, obj, "update")
	if err != nil {
		return err
	}

	o, err := iface.Update(ctx, u, opts)
	if err != nil {
		return err
	}

	err = fromUnstructured(o, obj)
	if err != nil {
		return err
	}

	return err
}

func (c *KubectlCmd) CreateOrUpdate(ctx context.Context, obj Object, f MutateFn) (OperationResult, error) {
	if err := c.Get(ctx, obj, metav1.GetOptions{}); err != nil {
		if !apierr.IsNotFound(err) {
			return OperationResultNone, err
		}

		err := mutate(f, ObjectKeyFromObject(obj), obj)
		if err != nil {
			return OperationResultNone, err
		}

		err = c.Create(ctx, obj, metav1.CreateOptions{
			FieldManager: "seaway",
		})
		if err != nil {
			return OperationResultNone, err
		}

		return OperationResultCreated, nil
	}

	existing, can := obj.DeepCopyObject().(Object)
	if !can {
		return OperationResultNone, fmt.Errorf("unable to cast object")
	}

	err := mutate(f, ObjectKeyFromObject(existing), obj)
	if err != nil {
		return OperationResultNone, err
	}

	err = c.Update(ctx, obj, metav1.UpdateOptions{
		FieldManager: "seaway",
	})
	if err != nil {
		return OperationResultNone, err
	}

	// TODO: Be better about merging the objects before sending
	// them to the API.  This is a hack to get around the fact that
	// I'm lazy and don't want to write a merge when non-nil function.
	// That being said, it's not like we are going to be pounding the
	// API with updates, so it's not really a priority.
	if reflect.DeepEqual(existing, obj) {
		return OperationResultNone, nil
	}

	return OperationResultUpdated, nil
}

// WaitForCondition blocks until specified condition to be met on the object or
// a timeout occurs.
func (c *KubectlCmd) WaitForCondition(ctx context.Context, obj Object, conditionString string, timeout time.Duration) error {
	resource, err := ResourceInterfaceFor(c.client, obj, "list")
	if err != nil {
		return err
	}

	kind := obj.GetObjectKind().GroupVersionKind().Kind
	wait := NewWaitCondition(obj.GetName(), kind, conditionString)
	return wait.WaitForCondition(ctx, WaitOptions{
		client:  resource,
		timeout: timeout,
	})
}

// Preserve some fields that are managed by the API. This is done to keep
// the resource from being updated when it is not necessary.  If managed fields
// are not preserved the API will write them back to the resource and bump the
// resource version and generation which makes the resource appear to be updated.
// We only handle the metadata fields here.  There are some resources that have
// spec fields that are managed by the API as well and they will show up as updated
// every time.  This is expected behavior and is present when using kubectl apply -k
// as well.
func PreserveManagedFields(source, target Object) {
	target.SetResourceVersion(source.GetResourceVersion())
	target.SetFinalizers(source.GetFinalizers())
	target.SetManagedFields(source.GetManagedFields())
	annotations := target.GetAnnotations()
	// TODO: I don't think I want this.  I just realized that we could have some
	// matches that we don't want to overwrite.
	// matcher := regexp.MustCompile(`^.*kubernetes.io\/.*$`)

	// if annotations == nil {
	// 	annotations = make(map[string]string)
	// }

	// for k, v := range source.GetAnnotations() {
	// 	if matcher.MatchString(k) {
	// 		annotations[k] = v
	// 	}
	// }
	target.SetAnnotations(annotations)
}

// ResourceInterfaceFor returns a new dynamic.ResourceInterface for the client.
func ResourceInterfaceFor(c *Client, obj Object, method string) (dynamic.ResourceInterface, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()

	dyn, err := c.Factory().DynamicClient()
	if err != nil {
		return nil, err
	}

	dc, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	resource, err := ServerResourcesForGroupVersionKind(dc, gvk, method)
	if err != nil {
		return nil, err
	}

	gvr := gvk.GroupVersion().WithResource(resource.Name)
	if resource.Namespaced {
		if obj.GetNamespace() == "" {
			obj.SetNamespace(corev1.NamespaceDefault)
		}
		return dyn.Resource(gvr).Namespace(obj.GetNamespace()), nil
	}
	return dyn.Resource(gvr), nil
}

// ServerResourcesForGroupVersionKind returns the APIResource for the provided GroupVersionKind.
func ServerResourcesForGroupVersionKind(dc discovery.CachedDiscoveryInterface, gvk schema.GroupVersionKind, verb string) (*metav1.APIResource, error) {
	resources, err := dc.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}

	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			if supportedVerb(&r, verb) {
				return &r, nil
			}

			return nil, apierr.NewMethodNotSupported(
				schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
				verb,
			)
		}
	}

	return nil, apierr.NewNotFound(
		schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
		"",
	)
}

func toUnstructured(obj Object, u *unstructured.Unstructured) error {
	switch o := obj.(type) { //nolint:gocritic
	case *unstructured.Unstructured:
		// TODO: I was thinking that the deepcopy was more performant than
		// the JSON marshaller and need to research whether that is true or
		// not.  To be honest, I don't think it matters if this is just used
		// for the client.
		o.DeepCopyInto(u)
		return nil
	}

	var data []byte
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return u.UnmarshalJSON(data)
}

func fromUnstructured(u *unstructured.Unstructured, obj Object) error {
	switch o := obj.(type) { //nolint:gocritic
	case *unstructured.Unstructured:
		u.DeepCopyInto(o)
		return nil
	}

	data, err := u.MarshalJSON()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}

func mutate(f MutateFn, key ObjectKey, obj Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := ObjectKeyFromObject(obj); key != newKey {
		// I dont think I'll handle this
		return nil
	}
	return nil
}

// supportedVerb returns true if the provided verb is supported by the APIResource.
func supportedVerb(apiResource *metav1.APIResource, verb string) bool {
	if verb == "" || verb == "*" {
		return true
	}

	for _, v := range apiResource.Verbs {
		if strings.EqualFold(v, verb) {
			return true
		}
	}
	return false
}
