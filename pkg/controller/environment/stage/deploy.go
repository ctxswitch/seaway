package stage

import (
	"context"
	"fmt"
	"reflect"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type operation string

const (
	OperationCreate operation = "create"
	OperationUpdate operation = "update"
	OperationNone   operation = "no changes"
)

type Deploy struct {
	observed *collector.ObservedState
	desired  *collector.DesiredState
	client.Client
}

func NewDeploy(client client.Client, collection *collector.Collection) *Deploy {
	return &Deploy{
		observed: collection.Observed,
		desired:  collection.Desired,
		Client:   client,
	}
}

func (d *Deploy) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx)

	if op, err := d.createOrUpdate(ctx, d.observed.Deployment, d.desired.Deployment); err != nil {
		status.Reason = fmt.Sprintf("Unable to %s deployment %s: %s", op, d.desired.Deployment.GetName(), err.Error())
		return v1beta1.EnvironmentStageDeployFailed, err
	} else {
		logger.V(5).Info("deployment", "operation", op)
	}

	// If we don't want a service, we won't have an ingress either, just return.
	if d.desired.Service == nil {
		return v1beta1.EnvironmentStageDeployVerify, nil
	}

	if op, err := d.createOrUpdate(ctx, d.observed.Service, d.desired.Service); err != nil {
		status.Reason = fmt.Sprintf("Unable to %s service %s: %s", op, d.desired.Service.GetName(), err.Error())
		return v1beta1.EnvironmentStageDeployFailed, err
	} else {
		logger.V(5).Info("service", "operation", op)
	}

	// If we don't have an ingress or it's not enabled, just return.
	if d.desired.Ingress == nil {
		return v1beta1.EnvironmentStageDeployVerify, nil
	}

	if op, err := d.createOrUpdate(ctx, d.observed.Ingress, d.desired.Ingress); err != nil {
		status.Reason = fmt.Sprintf("Unable to %s ingress %s: %s", op, d.desired.Ingress.GetName(), err.Error())
		return v1beta1.EnvironmentStageDeployFailed, err
	} else {
		logger.V(5).Info("ingress", "operation", op)
	}

	return v1beta1.EnvironmentStageDeployVerify, nil
}

func (d *Deploy) createOrUpdate(ctx context.Context, observed client.Object, desired client.Object) (operation, error) {
	logger := log.FromContext(ctx)
	if equality.Semantic.DeepEqual(observed, desired) {
		logger.V(3).Info("no changes detected", "kind", observed.GetObjectKind().GroupVersionKind().Kind, "object", observed.GetName())
		return OperationNone, nil
	}

	// Interfaces are crazy.
	if observed == nil || reflect.ValueOf(observed) == reflect.Zero(reflect.TypeOf(observed)) { //nolint:govet
		logger.V(3).Info("creating", "kind", desired.GetObjectKind().GroupVersionKind().Kind, "object", desired.GetName())
		return OperationCreate, d.Create(ctx, desired)
	}

	logger.V(3).Info("updating", "kind", desired.GetObjectKind().GroupVersionKind().Kind, "object", desired.GetName())
	return OperationUpdate, d.Update(ctx, desired)
}

var _ Stage = &Deploy{}
