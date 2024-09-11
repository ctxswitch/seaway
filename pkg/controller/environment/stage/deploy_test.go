package stage

import (
	"context"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestDeploy_DoNewEnvironmentAllComponents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_new_environment_1.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env.
	assert.NotNil(t, collection.Observed.Env)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.NoError(t, err)
	assert.NotNil(t, ingress)
}

func TestDeploy_DoNewEnvironmentOnlyDeploy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_new_environment_2.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env.
	assert.NotNil(t, collection.Observed.Env)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))
}

func TestDeploy_DoNewEnvironmentOnlyService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_new_environment_3.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env.
	assert.NotNil(t, collection.Observed.Env)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))
}

func TestDeploy_RemoveIngressIfNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_ingress_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env.
	assert.NotNil(t, collection.Observed.Env)
	assert.Nil(t, collection.Desired.Ingress)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))
}

func TestDeploy_RemoveServiceIfNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_ingress_service_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env since we are dependent on the observed state.
	// Probably not the best to couple the deps in the test like this.
	assert.NotNil(t, collection.Observed.Env)
	assert.Nil(t, collection.Desired.Ingress)
	assert.Nil(t, collection.Desired.Service)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))
}

// TestDeploy_DontRemoveIngressIfNotObserved tests that the service does not attempt
// to remove the ingress as part of it's cleanup if it was not observed.  Oops.
func TestDeploy_DontRemoveIngressIfNotObserved(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	mc.ApplyFixtureOrDie("deploy_service_no_ingress_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: mc,
	}
	err := sc.ObserveAndBuild(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	assert.NoError(t, err)
	// Sanity check to make sure we loaded the env since we are dependent on the observed state.
	// Probably not the best to couple the deps in the test like this.
	assert.NotNil(t, collection.Observed.Env)
	assert.NotNil(t, collection.Observed.Service)
	assert.Nil(t, collection.Observed.Ingress)
	assert.Nil(t, collection.Desired.Ingress)
	assert.Nil(t, collection.Desired.Service)

	d := NewDeploy(mc, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(ctx, status)
	assert.NoError(t, err)

	assert.Equal(t, v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	assert.NoError(t, err)
	assert.NotNil(t, deploy)

	var service corev1.Service
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = mc.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	assert.Error(t, err)
	assert.NoError(t, client.IgnoreNotFound(err))
}

func TestDeploy_createOrUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	mc := mock.NewClient().WithFixtureDirectory(fixtures).WithLogger(h.Logger())
	defer mc.Reset()

	d := NewDeploy(mc, &collector.Collection{})

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
			Annotations: map[string]string{
				"seaway.ctx.sh/revision": "1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	op, err := d.createOrUpdate(ctx, nil, &deploy)
	assert.NoError(t, err)
	assert.Equal(t, OperationCreate, op)

	updatedDeploy := deploy.DeepCopy()
	updatedDeploy.Spec.Template.Spec.Containers[0].Image = "nginx:1.19"

	op, err = d.createOrUpdate(ctx, &deploy, updatedDeploy)
	assert.NoError(t, err)
	assert.Equal(t, OperationUpdate, op)

	op, err = d.createOrUpdate(ctx, updatedDeploy, updatedDeploy)
	assert.NoError(t, err)
	assert.Equal(t, OperationNone, op)
}
