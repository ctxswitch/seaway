// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package environment

import (
	"context"
	"reflect"
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/stage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Controller struct {
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	RegistryURL      string
	RegistryNodePort int32
	client.Client
}

func SetupWithManager(mgr ctrl.Manager, regURL string, regNodePort int32) (err error) {
	c := &Controller{
		Scheme:           mgr.GetScheme(),
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("watch-controller"),
		RegistryURL:      regURL,
		RegistryNodePort: regNodePort,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Environment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(5).Info("reconciling development environment")

	env := &v1beta1.Environment{}
	err := c.Get(ctx, req.NamespacedName, env)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.V(5).Info("environment was deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to get environment")
		return ctrl.Result{}, err
	}

	v1beta1.Defaulted(env)
	// No finalizer support for now.

	// Change me up so if we aren't deployed yet, we call the reconciler - or if revision has changed mid deployment.

	if env.IsDeployed() {
		return ctrl.Result{}, nil
	}

	if env.HasDeviated() {
		logger.Info("environment has deviated, resetting", "revision", env.Spec.Revision)
		env.Status.Stage = v1beta1.EnvironmentCreateBuildJob
	}

	// If the job has failed, don't try to do anything with it.
	if env.HasFailed() {
		logger.Info("environment has failed, skipping", "stage", env.Status.Stage)
		return ctrl.Result{}, nil
	}

	status := env.Status.DeepCopy()
	current := status.Stage

	// TODO: Think about retry.
	next, err := c.reconcileStage(ctx, env, status, current)
	status.Stage = next

	c.updateStatus(ctx, env, status)

	if err != nil {
		logger.Error(err, "unable to reconcile environment")
		return ctrl.Result{}, err
	}

	if current == v1beta1.EnvironmentRevisionDeployed {
		logger.Info("environment deployed", "revision", env.Spec.Revision)
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, err
	}
}

func (c *Controller) reconcileStage(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus, condition v1beta1.EnvironmentCondition) (v1beta1.EnvironmentCondition, error) {
	reconciler := c.getReconciler(condition)
	next, err := reconciler.Do(ctx, env, status)
	if err != nil {
		return condition, err
	}

	return next, nil
}

func (c *Controller) getReconciler(current v1beta1.EnvironmentCondition) stage.Reconciler {
	switch current {
	case v1beta1.EnvironmentCreateBuildJob:
		return stage.NewBuild(c.Client, c.Scheme, c.RegistryURL)
	case v1beta1.EnvironmentWaitingForBuildJobToComplete:
		return stage.NewBuildWait(c.Client, c.Scheme, c.RegistryURL)
	case v1beta1.EnvironmentDeployingRevision:
		return stage.NewDeploy(c.Client, c.Scheme, c.RegistryNodePort)
	case v1beta1.EnvironmentWaitingForDeploymentToComplete:
		return stage.NewDeployWait(c.Client, c.Scheme)
	case v1beta1.EnvironmentRevisionDeployed:
		return stage.NewDeployed(c.Client, c.Scheme)
	default:
		return stage.NewBuild(c.Client, c.Scheme, c.RegistryURL)
	}
}

func (c *Controller) updateStatus(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) {
	logger := log.FromContext(ctx)
	logger.Info("updating environment status", "status", status)
	if !reflect.DeepEqual(env.Status.Stage, status) {
		status.DeepCopyInto(&env.Status)
		env.Status.LastUpdated = metav1.Now()
		err := c.Status().Update(ctx, env)
		if err != nil {
			logger.Error(err, "unable to update environment status")
		}
	}
}
