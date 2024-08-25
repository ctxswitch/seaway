package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployed struct {
	Scheme *runtime.Scheme
	client.Client
}

func NewDeployed(client client.Client, scheme *runtime.Scheme) *Deployed {
	return &Deployed{
		Client: client,
		Scheme: scheme,
	}
}

func (d *Deployed) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentCondition, error) {
	status.DeployedRevision = env.Spec.Revision
	return v1beta1.EnvironmentRevisionDeployed, nil
}
