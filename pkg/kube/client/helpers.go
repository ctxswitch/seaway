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
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"strings"
)

// ObjectKeyFromObject is a helper that returns an ObjectKey for the given
// Object.
func ObjectKeyFromObject(obj Object) ObjectKey {
	return ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func ConvertToUnstructured(obj Object) (*unstructured.Unstructured, error) {
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{Object: o}
	return u, nil

	// GVK?
	//
	//if gvk := obj.GetObjectKind().GroupVersionKind(); gvk.Empty() {}
}

func ConvertFromUnstructured(u *unstructured.Unstructured, obj Object) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
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
