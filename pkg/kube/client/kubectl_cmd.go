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
	"regexp"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// KubectlCmd is the kubernetes client that is used by the seactl tool.
type KubectlCmd struct {
	dc     dynamic.Interface
	client *Client
}

func NewKubectlCmd(ns, context string) (*KubectlCmd, error) {
	c, err := NewClient(ns, context)
	if err != nil {
		return nil, err
	}

	dc, err := c.Factory().DynamicClient()
	if err != nil {
		return nil, err
	}

	return &KubectlCmd{
		dc:     dc,
		client: c,
	}, nil
}

func (c *KubectlCmd) Get(ctx context.Context, obj Object, opts metav1.GetOptions) error {
	u := &unstructured.Unstructured{}
	if err := toUnstructured(obj, u); err != nil {
		return err
	}

	iface, err := c.client.ResourceInterfaceFor(obj, "get")
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

	iface, err := c.client.ResourceInterfaceFor(obj, "delete")
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

	iface, err := c.client.ResourceInterfaceFor(obj, "create")
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

	iface, err := c.client.ResourceInterfaceFor(obj, "update")
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
		if !errors.IsNotFound(err) {
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
	matcher := regexp.MustCompile(`^.*kubernetes.io\/.*$`)

	if annotations == nil {
		annotations = make(map[string]string)
	}

	for k, v := range source.GetAnnotations() {
		if matcher.MatchString(k) {
			annotations[k] = v
		}
	}
	target.SetAnnotations(annotations)
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
