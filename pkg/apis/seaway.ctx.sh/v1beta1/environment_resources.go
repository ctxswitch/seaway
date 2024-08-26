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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CoreV1ResourceRequirements returns a corev1.ResourceRequirements for use when building
// container objects.
func (r EnvironmentResources) CoreV1ResourceRequirements() corev1.ResourceRequirements {
	req := corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}

	for k, v := range r {
		req.Requests[k] = v
		req.Limits[k] = v
	}

	return req
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.  The normal yaml parser does
// not support the conversion of strings into the resource.Quantity type.  This function
// allows for the conversion and validation. This is used exclusively for the manifest
// loading process in the client.
func (r *EnvironmentResources) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type EnvironmentResourcesRaw map[corev1.ResourceName]string

	raw := EnvironmentResourcesRaw{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	out := EnvironmentResources{}
	for k, v := range raw {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return err
		}
		out[k] = q
	}

	*r = out
	return nil
}
