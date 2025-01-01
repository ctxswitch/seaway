package stage

import (
	"context"
	"path/filepath"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type DeployTestSuite struct {
	client *mock.Client
	suite.Suite
}

func (s *DeployTestSuite) SetupTest() {
	logger := zap.New(zap.UseDevMode(true), zap.Level(zapcore.Level(-8)))
	log.SetLogger(logger)

	s.client = mock.NewClient().
		WithLogger(logger).
		WithFixtureDirectory(filepath.Join("..", "..", "..", "..", "fixtures"))

	s.client.ApplyFixtureOrDie("shared", "required.yaml")
}

func (s *DeployTestSuite) TearDownTest() {
	s.client.Reset()
}

func TestDeployTestSuite(t *testing.T) {
	suite.Run(t, new(DeployTestSuite))
}

func (s *DeployTestSuite) TestDeploy_DoNewEnvironmentAllComponents() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_new_environment_1.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Assert().NoError(err)
	s.Assert().NotNil(service)

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Assert().NoError(err)
	s.Assert().NotNil(ingress)
}

func (s *DeployTestSuite) TestDeploy_DoNewEnvironmentOnlyDeploy() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_new_environment_2.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))
}

func (s *DeployTestSuite) TestDeploy_DoNewEnvironmentOnlyService() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_new_environment_3.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Assert().NoError(err)
	s.Assert().NotNil(service)

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))
}

func (s *DeployTestSuite) TestDeploy_RemoveIngressIfNil() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_ingress_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	s.Nil(collection.Desired.Ingress)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Assert().NoError(err)
	s.Assert().NotNil(service)

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))
}

func (s *DeployTestSuite) TestDeploy_RemoveServiceIfNil() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_ingress_service_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	s.Nil(collection.Desired.Ingress)
	s.Nil(collection.Desired.Service)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))
}

// TestDeploy_DontRemoveIngressIfNotObserved tests that the service does not attempt
// to remove the ingress as part of it's cleanup if it was not observed.  Oops.
func (s *DeployTestSuite) TestDeploy_DontRemoveIngressIfNotObserved() {
	s.client.ApplyFixtureOrDie("controller_environment_stage", "deploy_service_no_ingress_cleanup.yaml")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client:           s.client,
		BuilderNamespace: "seaway-build",
	}
	err := sc.ObserveAndBuild(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}, &collection)
	s.Assert().NoError(err)
	// Sanity check to make sure we loaded the required components.
	s.Assert().NotNil(collection.Observed.Env)
	s.Assert().NotNil(collection.Observed.Config)
	s.Assert().NotNil(collection.Observed.StorageCredentials)

	s.Assert().NotNil(collection.Observed.Service)
	s.Nil(collection.Observed.Ingress)
	s.Nil(collection.Desired.Ingress)
	s.Nil(collection.Desired.Service)

	d := NewDeploy(s.client, &collection)
	status := collection.Observed.Env.Status.DeepCopy()

	stage, err := d.Do(context.TODO(), status)
	s.Assert().NoError(err)

	s.Assert().Equal(v1beta1.EnvironmentStageDeployVerify, stage)

	var deploy appsv1.Deployment
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &deploy)
	s.Assert().NoError(err)
	s.Assert().NotNil(deploy)

	var service corev1.Service
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &service)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))

	var ingress networkingv1.Ingress
	err = s.client.Get(context.TODO(), types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}, &ingress)
	s.Error(err)
	s.NoError(client.IgnoreNotFound(err))
}

func (s *DeployTestSuite) TestDeploy_createOrUpdate() {
	d := NewDeploy(s.client, &collector.Collection{})

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
	op, err := d.createOrUpdate(context.TODO(), nil, &deploy)
	s.Assert().NoError(err)
	s.Assert().Equal(OperationCreate, op)

	updatedDeploy := deploy.DeepCopy()
	updatedDeploy.Spec.Template.Spec.Containers[0].Image = "nginx:1.19"

	op, err = d.createOrUpdate(context.TODO(), &deploy, updatedDeploy)
	s.Assert().NoError(err)
	s.Assert().Equal(OperationUpdate, op)

	op, err = d.createOrUpdate(context.TODO(), updatedDeploy, updatedDeploy)
	s.Assert().NoError(err)
	s.Assert().Equal(OperationNone, op)
}
