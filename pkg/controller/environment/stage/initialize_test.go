package stage

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type expected struct {
	Stage    v1beta1.EnvironmentStage
	Expected string
	Current  string
}

func TestInitialize(t *testing.T) {
	var tests = []struct {
		desc        string
		environment *v1beta1.Environment
		expected    expected
	}{
		{
			desc: "new environment",
			environment: &v1beta1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1beta1.EnvironmentSpec{
					Revision: "1",
				},
				Status: v1beta1.EnvironmentStatus{},
			},
			expected: expected{
				Stage:    v1beta1.EnvironmentStageBuildImage,
				Expected: "1",
				Current:  "",
			},
		},
		{
			desc: "partially deployed environment",
			environment: &v1beta1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1beta1.EnvironmentSpec{
					Revision: "2",
				},
				Status: v1beta1.EnvironmentStatus{
					ExpectedRevision: "1",
					DeployedRevision: "",
					Stage:            v1beta1.EnvironmentStageBuildImageVerify,
				},
			},
			expected: expected{
				Stage:    v1beta1.EnvironmentStageBuildImage,
				Expected: "2",
				Current:  "",
			},
		},
		{
			desc: "deployed environment",
			environment: &v1beta1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1beta1.EnvironmentSpec{
					Revision: "2",
				},
				Status: v1beta1.EnvironmentStatus{
					ExpectedRevision: "2",
					DeployedRevision: "1",
					Stage:            v1beta1.EnvironmentStageDeployed,
				},
			},
			expected: expected{
				Stage:    v1beta1.EnvironmentStageBuildImage,
				Expected: "2",
				Current:  "1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			client := mock.NewClient()
			stage := NewInitialize(client, &collector.Collection{
				Observed: &collector.ObservedState{
					Env: test.environment,
					StorageCredentials: &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "storage-credentials",
							Namespace: "seaway-build",
						},
					},
					BuilderNamespace: &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "seaway-build",
						},
					},
				},
			})
			status := test.environment.Status.DeepCopy()
			next, err := stage.Do(context.TODO(), status)
			assert.NoError(t, err)
			assert.Equal(t, test.expected.Stage, next)
			assert.Equal(t, test.expected.Expected, status.ExpectedRevision)
			assert.Equal(t, test.expected.Current, status.DeployedRevision)
		})
	}
}
