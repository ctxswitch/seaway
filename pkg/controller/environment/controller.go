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

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Controller struct {
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	client.Client
}

func SetupWithManager(mgr ctrl.Manager) (err error) {
	c := &Controller{
		Scheme:   mgr.GetScheme(),
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("watch-controller"),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Environment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=environments/finalizers,verbs=update
// +kubebuilder:rbac:groups=seaway.ctx.sh,resources=seawayconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(5).Info("reconciling development environment")

	var collection collector.Collection

	sc := &collector.StateCollector{
		Client: c.Client,
		Scheme: c.Scheme,
	}
	if err := sc.ObserveAndBuild(ctx, req, &collection); err != nil {
		return ctrl.Result{}, err
	}

	if collection.Observed.Env == nil {
		logger.Info("environment was deleted")
		return ctrl.Result{}, nil
	}

	handler := &Handler{
		collection: &collection,
		client:     c.Client,
	}

	return handler.reconcile(ctx)
}
