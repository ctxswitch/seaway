package stage

import (
	"context"
	"errors"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BuildImageWait struct {
	observed *collector.ObservedState
	desired  *collector.DesiredState
	client.Client
}

func NewBuildImageWait(client client.Client, collection *collector.Collection) *BuildImageWait {
	return &BuildImageWait{
		observed: collection.Observed,
		desired:  collection.Desired,
		Client:   client,
	}
}

func (b *BuildImageWait) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx)
	logger.V(4).Info("waiting for build job to complete")

	job := b.observed.Job
	if job.Status.Active > 0 && job.Status.Failed > 0 {
		return v1beta1.EnvironmentStageBuildImageFailing, nil
	} else {
		if job.Status.CompletionTime != nil {
			return v1beta1.EnvironmentStageBuildImageVerify, nil
		}

		if len(job.Status.Conditions) > 0 {
			next := v1beta1.EnvironmentStageBuildImageFailed
			return next, errors.New("build failed")
		}
	}

	return v1beta1.EnvironmentStageBuildImageWait, nil
}

var _ Stage = &BuildImageWait{}
