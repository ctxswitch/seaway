package stage

import (
	"context"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployVerifyExpected struct {
	revision string
	stage    v1beta1.EnvironmentStage
}

func TestDeployVerify(t *testing.T) {
	var tests = []struct {
		name      string
		replicas  int32
		available int32
		revision  string
		expected  DeployVerifyExpected
	}{
		{
			name:      "no available replicas",
			replicas:  1,
			available: 0,
			revision:  "2",
			expected: DeployVerifyExpected{
				revision: "",
				stage:    v1beta1.EnvironmentStageDeployVerify,
			},
		},
		{
			name:      "all replicas available",
			replicas:  1,
			available: 1,
			revision:  "2",
			expected: DeployVerifyExpected{
				revision: "2",
				stage:    v1beta1.EnvironmentStageDeployed,
			},
		},
		{
			name:      "replicas partially available",
			replicas:  2,
			available: 1,
			revision:  "2",
			expected: DeployVerifyExpected{
				revision: "",
				stage:    v1beta1.EnvironmentStageDeployVerify,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := collector.Collection{
				Observed: &collector.ObservedState{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &tt.replicas,
						},
						Status: appsv1.DeploymentStatus{
							AvailableReplicas: tt.available,
						},
					},
					Env: &v1beta1.Environment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
						Spec: v1beta1.EnvironmentSpec{
							Revision: tt.revision,
						},
						Status: v1beta1.EnvironmentStatus{
							ExpectedRevision: "2",
							DeployedRevision: "1",
						},
					},
				},
			}

			d := NewDeployVerify(nil, &collection)
			status := v1beta1.EnvironmentStatus{}
			stage, err := d.Do(context.TODO(), &status)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.stage, stage)
			assert.Equal(t, tt.expected.revision, status.DeployedRevision)
		})
	}
}
