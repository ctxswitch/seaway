package stage

import (
	"context"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/registry"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockRegistry struct {
	HasTagRespose bool
}

func NewMockRegistry() registry.API {
	return &MockRegistry{}
}

func (m *MockRegistry) WithRegistry(reg string) registry.API {
	return m
}

func (m *MockRegistry) SetHasTag(response bool) {
	m.HasTagRespose = response
}

func (m *MockRegistry) HasTag(name, tag string) (bool, error) {
	return m.HasTagRespose, nil
}

var _ registry.API = &MockRegistry{}

func TestBuildImageVerify(t *testing.T) {
	var tests = []struct {
		name     string
		hasTag   bool
		expected v1beta1.EnvironmentStage
	}{
		{
			name:     "tag exists",
			hasTag:   true,
			expected: v1beta1.EnvironmentStageDeploy,
		},
		{
			name:     "tag does not exist",
			hasTag:   false,
			expected: v1beta1.EnvironmentStageBuildImageFailed,
		},
	}

	for _, tt := range tests {
		collection := &collector.Collection{
			Observed: &collector.ObservedState{
				Config: &v1beta1.EnvironmentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: v1beta1.DefaultControllerNamespace,
					},
					Spec: v1beta1.EnvironmentConfigSpec{
						EnvironmentConfigRegistrySpec: v1beta1.EnvironmentConfigRegistrySpec{
							URL: "http://registry:5000",
						},
					},
				},
				Env: &v1beta1.Environment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: v1beta1.EnvironmentSpec{
						Revision: "1",
					},
				},
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := NewMockRegistry()
			mockRegistry.(*MockRegistry).SetHasTag(tt.hasTag)

			b := NewBuildImageVerify(nil, collection).WithRegistry(mockRegistry)
			stage, err := b.Do(context.TODO(), &v1beta1.EnvironmentStatus{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, stage)
		})
	}
}
