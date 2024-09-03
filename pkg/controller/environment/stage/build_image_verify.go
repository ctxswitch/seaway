package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"ctx.sh/seaway/pkg/registry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BuildImageVerify struct {
	observed *collector.ObservedState
	desired  *collector.DesiredState
	registry registry.API
	client.Client
}

func NewBuildImageVerify(client client.Client, collection *collector.Collection, reg registry.API) *BuildImageVerify {
	return &BuildImageVerify{
		observed: collection.Observed,
		desired:  collection.Desired,
		registry: reg,
		Client:   client,
	}
}

func (b *BuildImageVerify) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	env := b.observed.Env

	if ok, err := b.registry.HasTag(env.GetName(), env.GetRevision()); !ok || err != nil {
		return v1beta1.EnvironmentStageBuildImageFailed, err
	}

	return v1beta1.EnvironmentStageDeploy, nil
}

var _ Stage = &BuildImageVerify{}
