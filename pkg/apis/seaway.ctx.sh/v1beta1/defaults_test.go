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
			Network: &EnvironmentNetwork{
				Service: &EnvironmentService{
					Enabled: false,
				},
				Ingress: nil,
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

func TestDefaulted_ServiceEnabled(t *testing.T) {
	obj := &Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: EnvironmentSpec{
			Network: &EnvironmentNetwork{
				Service: &EnvironmentService{
					Enabled: true,
				},
				Ingress: nil,
			},
		},
	}

	expected := &EnvironmentNetwork{
		Service: &EnvironmentService{
			Enabled:     true,
			Annotations: map[string]string{},
			Type:        corev1.ServiceTypeClusterIP,
			Ports: []EnvironmentPort{
				{
					Name:     "http",
					Port:     9000,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
		Ingress: &EnvironmentIngress{
			Enabled: false,
		},
	}

	Defaulted(obj)
	assert.Equal(t, expected, obj.Spec.Network)
}

func TestDefaulted_IngressEnabled(t *testing.T) {
	obj := &Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: EnvironmentSpec{
			Network: &EnvironmentNetwork{
				Service: &EnvironmentService{
					Enabled: true,
				},
				Ingress: &EnvironmentIngress{
					Enabled: true,
				},
			},
		},
	}

	expected := &EnvironmentNetwork{
		Service: &EnvironmentService{
			Enabled:     true,
			Annotations: map[string]string{},
			Type:        corev1.ServiceTypeClusterIP,
			Ports: []EnvironmentPort{
				{
					Name:     "http",
					Port:     9000,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
		Ingress: &EnvironmentIngress{
			Enabled:     true,
			Annotations: map[string]string{},
			Port:        ptr.To(int32(9000)),
		},
	}

	Defaulted(obj)
	assert.Equal(t, expected, obj.Spec.Network)
}
