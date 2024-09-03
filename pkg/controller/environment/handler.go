package environment

import (
	"context"
	"reflect"
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/controller/environment/stage"
	"ctx.sh/seaway/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ReconcilerFunc func(context.Context, *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error)

type Handler struct {
	client     client.Client
	collection *collector.Collection
	registry   registry.API
}

func (h *Handler) reconcile(ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("revision", h.collection.Observed.Env.Spec.Revision)
	// ctx = log.IntoContext(ctx, logger)

	logger.V(5).Info("handling reconciliation for revision")

	if h.collection.Observed.UserSecret == nil {
		logger.Error(nil, "unable to reconcile environment without user secrets")
		return ctrl.Result{}, nil
	}

	env := h.collection.Observed.Env

	// There's probably a cleaner way to handle the initial checks here.

	// If there was a build failure we stop.
	// If there was a deploy failure we can try to redeploy.

	switch {
	case env.Status.Stage == v1beta1.EnvironmentStageDeployFailed:
		logger.Info("environment has failed to deploy, skipping reconcile")
		return ctrl.Result{}, nil
	case env.Status.Stage == v1beta1.EnvironmentStageBuildImageFailed:
		logger.Info("environment has failed to build, skipping reconcile")
		return ctrl.Result{}, nil
	case env.HasDeviated():
		logger.Info("environment been updated, redeploying")
		env.Status.Stage = v1beta1.EnvironmentStageInitialize
	case env.IsDeployed():
		logger.V(5).Info("handling reconciliation for existing revision")
		// env.Status.Stage = v1beta1.EnvironmentStageDeploy
		// TODO: Think about how we can handle a redeploy if only the env resources
		// have changed and not the image itself.  Might want to add a manifest hash.
		return ctrl.Result{}, nil
	}

	status := env.Status.DeepCopy()
	stage := status.Stage

	// TODO: Think about retry also need to wrap a timeout from the last status update.
	s := h.getStage(stage)
	next, err := s.Do(ctx, status)
	status.Stage = next

	h.updateStatus(ctx, env, status)

	if err != nil {
		logger.Error(err, "unable to reconcile environment", "next", next, "status", status)
		return ctrl.Result{}, err
	}

	if next == v1beta1.EnvironmentStageDeployed {
		logger.Info("revision has been deployed")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: time.Second}, nil
}

func (h *Handler) getStage(current v1beta1.EnvironmentStage) stage.Stage {
	switch current {
	// Initialize
	case v1beta1.EnvironmentStageInitialize:
		return stage.NewInitialize(h.client, h.collection)
	case v1beta1.EnvironmentStageBuildImage:
		return stage.NewBuildImage(h.client, h.collection)
	case v1beta1.EnvironmentStageBuildImageWait:
		return stage.NewBuildImageWait(h.client, h.collection)
	case v1beta1.EnvironmentStageBuildImageVerify:
		return stage.NewBuildImageVerify(h.client, h.collection, h.registry)
	case v1beta1.EnvironmentStageDeploy:
		return stage.NewDeploy(h.client, h.collection)
	case v1beta1.EnvironmentStageDeployVerify:
		return stage.NewDeployVerify(h.client, h.collection)
	default:
		return stage.NewError()
	}
}

func (h *Handler) updateStatus(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) {
	logger := log.FromContext(ctx)
	logger.Info("updating environment status", "status", status)

	if !reflect.DeepEqual(env.Status.Stage, status) {
		status.DeepCopyInto(&env.Status)
		env.Status.LastUpdated = metav1.Now()
		err := h.client.Status().Update(ctx, env)
		if err != nil {
			logger.Error(err, "unable to update environment status")
		}
	}
}
