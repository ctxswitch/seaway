package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Initialize struct {
	observed *collector.ObservedState
	client.Client
}

func NewInitialize(client client.Client, collection *collector.Collection) *Initialize {
	return &Initialize{
		observed: collection.Observed,
		Client:   client,
	}
}

func (i *Initialize) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	status.ExpectedRevision = i.observed.Env.GetRevision()

	// Ensure that the builder namespace exists.
	if i.observed.BuilderNamespace == nil {
		err := i.Create(ctx, i.observed.BuilderNamespace)
		if err != nil {
			return v1beta1.EnvironmentStageInitialize, err
		}
	}

	// TODO: ensure that the build secret exists.

	return v1beta1.EnvironmentStageBuildImage, nil
}

var _ Stage = &Initialize{}
