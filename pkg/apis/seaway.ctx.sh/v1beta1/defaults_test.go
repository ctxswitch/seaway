package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestDefaulted(t *testing.T) {
	expected := &Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: EnvironmentSpec{
			Args: []string{},
			Build: &EnvironmentBuild{
				Dockerfile: ptr.To(DefaultDockerfile),
				Image:      ptr.To(DefaultBuildImage),
				Platform:   ptr.To(DefaultPlatform),
				Include:    []string{},
				Exclude:    []string{},
			},
			Command:       nil,
			Lifecycle:     nil,
			LivenessProbe: nil,
			Networking: &EnvironmentNetworking{
				Ports: []EnvironmentPort{},
				Ingress: &EnvironmentIngress{
					Annotations: map[string]string{},
					Enabled:     false,
				},
			},
			ReadinessProbe: nil,
			Replicas:       ptr.To(int32(DefaultReplicas)),
			Resources:      EnvironmentResources{},
			Store: &EnvironmentStore{
				Bucket:         ptr.To(DefaultBucket),
				Prefix:         ptr.To(DefaultPrefix),
				Region:         ptr.To(DefaultRegion),
				Endpoint:       ptr.To(DefaultEndpoint),
				ForcePathStyle: ptr.To(DefaultForcePathStyle),
				LocalPort:      nil,
			},
			StartupProbe: nil,
			Vars: &EnvironmentVars{
				Env:     []corev1.EnvVar{},
				EnvFrom: []corev1.EnvFromSource{},
			},
			WorkingDir: "",
		},
	}

	obj := &Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	Defaulted(obj)
	assert.Equal(t, expected, obj)
}
