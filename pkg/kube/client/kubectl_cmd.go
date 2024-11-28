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
	"reflect"
	"time"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	u, err := ConvertToUnstructured(obj)
	if err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, u, "get")
	if err != nil {
		return err
	}

	u, err = iface.Get(ctx, u.GetName(), opts)
	if err != nil {
		return err
	}

	return ConvertFromUnstructured(u, obj)
}

func (c *KubectlCmd) Delete(ctx context.Context, obj Object, opts metav1.DeleteOptions) error {
	iface, err := ResourceInterfaceFor(c.client, obj, "delete")
	if err != nil {
		return err
	}

	if opts.PropagationPolicy == nil {
		propagationPolicy := metav1.DeletePropagationForeground
		opts = metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}
	}

	return iface.Delete(ctx, obj.GetName(), opts)
}

func (c *KubectlCmd) Create(ctx context.Context, obj Object, opts metav1.CreateOptions) error {
	u, err := ConvertToUnstructured(obj)
	if err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, u, "create")
	if err != nil {
		return err
	}

	u, err = iface.Create(ctx, u, opts)
	if err != nil {
		return err
	}

	return ConvertFromUnstructured(u, obj)
}

func (c *KubectlCmd) Update(ctx context.Context, obj Object, opts metav1.UpdateOptions) error {
	u, err := ConvertToUnstructured(obj)
	if err != nil {
		return err
	}

	iface, err := ResourceInterfaceFor(c.client, u, "update")
	if err != nil {
		return err
	}

	u, err = iface.Update(ctx, u, opts)
	if err != nil {
		return err
	}

	return ConvertFromUnstructured(u, obj)
}

func (c *KubectlCmd) CreateOrUpdate(ctx context.Context, obj Object, fn MutateFn) (OperationResult, error) {
	err := c.Get(ctx, obj, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return OperationResultNone, err
		}

		err = fn()
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

	before := obj.DeepCopyObject()
	err = fn()
	if err != nil {
		return OperationResultNone, err
	}

	err = c.Update(ctx, obj, metav1.UpdateOptions{
		FieldManager: "seaway",
	})
	if err != nil {
		return OperationResultNone, err
	}

	if reflect.DeepEqual(obj, before) {
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
