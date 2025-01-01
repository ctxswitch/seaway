package stage

import (
	"context"
	"path/filepath"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestBuildImage_NoChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().
		WithFixtureDirectory(filepath.Join("..", "..", "..", "..", "fixtures")).
		WithLogger(h.Logger())

	defer mc.Reset()

	mc.ApplyFixtureOrDie("shared", "required.yaml")

	collection := collector.Collection{
		Observed: &collector.ObservedState{
			BuilderNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "seaway-build",
				},
			},
			Job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-build",
					Namespace: "seaway-build",
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
		Desired: &collector.DesiredState{
			Job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-build",
					Namespace: "seaway-build",
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

	b := NewBuildImage(mc, &collection)
	status := &v1beta1.EnvironmentStatus{}
	stage, err := b.Do(context.TODO(), status)
	assert.NoError(t, err)
	assert.Equal(t, v1beta1.EnvironmentStageDeploy, stage)
}

func TestBuildImage_ChangeNoJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	collection := collector.Collection{
		Observed: &collector.ObservedState{
			BuilderNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "seaway-builder",
				},
			},
		},
		Desired: &collector.DesiredState{
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

	b := NewBuildImage(mc, &collection)
	status := &v1beta1.EnvironmentStatus{}
	stage, err := b.Do(context.TODO(), status)
	assert.NoError(t, err)
	assert.Equal(t, v1beta1.EnvironmentStageBuildImageWait, stage)

	var job batchv1.Job
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test-build",
		Namespace: "default",
	}, &job)
	assert.NoError(t, err)
	assert.Equal(t, collection.Desired.Job, &job)
}

func TestBuildImage_JobDeletedAfterObserve(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	collection := collector.Collection{
		Observed: &collector.ObservedState{
			BuilderNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "seaway-builder",
				},
			},
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
		Desired: &collector.DesiredState{
			Job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-build",
					Namespace: "default",
					Annotations: map[string]string{
						"seaway.ctx.sh/revision": "2",
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

	b := NewBuildImage(mc, &collection)
	status := &v1beta1.EnvironmentStatus{}
	stage, err := b.Do(context.TODO(), status)
	// We would expect an error if the existing job was not deleted.
	assert.NoError(t, err)
	assert.Equal(t, v1beta1.EnvironmentStageBuildImageWait, stage)

	var job batchv1.Job
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test-build",
		Namespace: "default",
	}, &job)
	assert.NoError(t, err)
	assert.Equal(t, collection.Desired.Job, &job)
}

func TestBuildImage_PreviousJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	collection := collector.Collection{
		Observed: &collector.ObservedState{
			BuilderNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "seaway-builder",
				},
			},
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
		Desired: &collector.DesiredState{
			Job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-build",
					Namespace: "default",
					Annotations: map[string]string{
						"seaway.ctx.sh/revision": "2",
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

	err := mc.Create(ctx, collection.Observed.Job)
	assert.NoError(t, err)

	b := NewBuildImage(mc, &collection)
	status := &v1beta1.EnvironmentStatus{}
	stage, err := b.Do(context.TODO(), status)
	assert.NoError(t, err)
	assert.Equal(t, v1beta1.EnvironmentStageBuildImageWait, stage)

	var job batchv1.Job
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test-build",
		Namespace: "default",
	}, &job)
	assert.NoError(t, err)
	assert.Equal(t, collection.Desired.Job, &job)
}
