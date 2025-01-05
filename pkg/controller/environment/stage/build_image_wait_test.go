package stage

import (
	"context"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestBuildImageWait(t *testing.T) {
	var tests = []struct {
		name     string
		status   batchv1.JobStatus
		expected v1beta1.EnvironmentStage
	}{
		{
			name: "job started, but not yet active",
			status: batchv1.JobStatus{
				Active:         0,
				Ready:          nil,
				Failed:         0,
				CompletionTime: nil,
				Conditions:     []batchv1.JobCondition{},
			},
			expected: v1beta1.EnvironmentStageBuildImageWait,
		},
		{
			name: "job started, is active but not ready",
			status: batchv1.JobStatus{
				Active:         1,
				Ready:          nil,
				Failed:         0,
				CompletionTime: nil,
				Conditions:     []batchv1.JobCondition{},
			},
			expected: v1beta1.EnvironmentStageBuildImageWait,
		},
		{
			name: "job started, is active and ready",
			status: batchv1.JobStatus{
				Active:         1,
				Ready:          ptr.To(int32(1)),
				Failed:         0,
				CompletionTime: nil,
				Conditions:     []batchv1.JobCondition{},
			},
			expected: v1beta1.EnvironmentStageBuildImageWait,
		},
		{
			name: "job started, is active and ready",
			status: batchv1.JobStatus{
				Active:         0,
				Ready:          nil,
				Failed:         0,
				CompletionTime: ptr.To(metav1.Now()),
				Conditions:     []batchv1.JobCondition{},
			},
			expected: v1beta1.EnvironmentStageBuildImageVerify,
		},
		{
			name: "job started, is active and failing",
			status: batchv1.JobStatus{
				Active: 1,
				// Ready does not matter in any of the checks if we have failed runs
				Ready:          nil,
				Failed:         1,
				CompletionTime: nil,
				Conditions:     []batchv1.JobCondition{},
			},
			expected: v1beta1.EnvironmentStageBuildImageFailing,
		},
		{
			name: "job has failed",
			status: batchv1.JobStatus{
				Active:         0,
				Ready:          nil,
				Failed:         0,
				CompletionTime: nil,
				Conditions: []batchv1.JobCondition{
					{
						Type:   batchv1.JobFailed,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: v1beta1.EnvironmentStageBuildImageFailed,
		},
	}

	collection := collector.Collection{
		Observed: &collector.ObservedState{
			Job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-build",
					Namespace: "default",
					Annotations: map[string]string{
						"seaway.ctx.sh/revision": "1",
					},
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "build",
									Image: v1beta1.DefaultBuildImage,
								},
							},
						},
					},
				},
			},
		},
	}

	mc := mock.NewClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection.Observed.Job.Status = tt.status
			bv := NewBuildImageWait(mc, &collection)
			stage, err := bv.Do(context.TODO(), &v1beta1.EnvironmentStatus{})

			if tt.expected == v1beta1.EnvironmentStageBuildImageFailed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, stage)
		})
	}
}

func TestBuildImageWait_NilJob(t *testing.T) {
	collection := collector.Collection{
		Observed: &collector.ObservedState{
			Job: nil,
		},
	}

	mc := mock.NewClient()
	bv := NewBuildImageWait(mc, &collection)
	stage, err := bv.Do(context.TODO(), &v1beta1.EnvironmentStatus{})
	assert.Error(t, err)
	assert.Equal(t, v1beta1.EnvironmentStageBuildImageFailed, stage)
}
