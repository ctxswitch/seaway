package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type DeployWait struct {
	Scheme *runtime.Scheme
	client.Client
}

func NewDeployWait(client client.Client, scheme *runtime.Scheme) *DeployWait {
	return &DeployWait{
		Client: client,
		Scheme: scheme,
	}
}

func (d *DeployWait) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentCondition, error) {
	logger := log.FromContext(ctx)
	logger.Info("waiting for deployment to complete")

	// TODO: Check for timeout and fail.

	deploy := GetEnvironmentDeployment(env, d.Scheme)
	err := d.Get(ctx, client.ObjectKeyFromObject(&deploy), &deploy)
	if err != nil {
		logger.Error(err, "unable to get deployment")
		return v1beta1.EnvironmentDeploymentFailed, err
	}

	if deploy.Status.AvailableReplicas < *deploy.Spec.Replicas {
		return v1beta1.EnvironmentWaitingForDeploymentToComplete, nil
	}

	status.DeployedRevision = env.Spec.Revision
	return v1beta1.EnvironmentRevisionDeployed, nil
}
