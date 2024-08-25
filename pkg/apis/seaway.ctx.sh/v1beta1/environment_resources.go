package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

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
// allows for the conversion and validation.
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
