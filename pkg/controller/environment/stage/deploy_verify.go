package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeployVerify struct {
	observed *collector.ObservedState
	desired  *collector.DesiredState
	Client   client.Client
}

func NewDeployVerify(client client.Client, collection *collector.Collection) *DeployVerify {
	return &DeployVerify{
		observed: collection.Observed,
		desired:  collection.Desired,
		Client:   client,
	}
}

func (d *DeployVerify) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	deploy := d.observed.Deployment
	env := d.observed.Env

	if deploy.Status.AvailableReplicas < *deploy.Spec.Replicas {
		return v1beta1.EnvironmentStageDeployVerify, nil
	}

	status.DeployedRevision = env.GetRevision()
	return v1beta1.EnvironmentStageDeployed, nil
}

var _ Stage = &DeployVerify{}
