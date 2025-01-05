package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller/environment/collector"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BuildImage struct {
	observed *collector.ObservedState
	desired  *collector.DesiredState
	client.Client
}

func NewBuildImage(client client.Client, collection *collector.Collection) *BuildImage {
	return &BuildImage{
		observed: collection.Observed,
		desired:  collection.Desired,
		Client:   client,
	}
}

func (b *BuildImage) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx)
	logger.V(4).Info("building image", "job", b.desired.Job)

	if equality.Semantic.DeepEqual(b.observed.Job, b.desired.Job) {
		logger.V(4).Info("job has not changed, skipping creation")
		return v1beta1.EnvironmentStageDeploy, nil
	}

	if b.observed.Job != nil {
		logger.V(4).Info("deleting old job", "job", b.observed.Job.Name)
		// if the job has changed, delete the old job
		if err := b.Delete(ctx, b.observed.Job, &client.DeleteOptions{
			// If the propegation policy is not set, the pods will not be deleted
			// when the job is deleted.  For some reason it defaults to Orphan, which
			// seems like an odd default to me.
			PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
		}); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return v1beta1.EnvironmentStageBuildImageFailed, err
			}
			// If the job was observed but deleted before we we could delete it, we
			// don't fail the build since we are in the correct state.
			logger.V(4).Info("job was deleted before we could delete it", "job", b.observed.Job.Name)
		}
	}

	logger.V(4).Info("creating new job", "job", b.desired.Job.ObjectMeta)
	err := b.Create(ctx, b.desired.Job)
	if err != nil {
		return v1beta1.EnvironmentStageBuildImageFailed, err
	}

	return v1beta1.EnvironmentStageBuildImageWait, nil
}

var _ Stage = &BuildImage{}
