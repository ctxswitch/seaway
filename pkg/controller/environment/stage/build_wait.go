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

package stage

import (
	"context"
	"errors"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BuildWait is the stage that waits for the build job to complete.
type BuildWait struct {
	Scheme      *runtime.Scheme
	RegistryURL string
	client.Client
}

// NewBuildWait creates a new BuildWait stage.
func NewBuildWait(client client.Client, scheme *runtime.Scheme, registryURL string) *BuildWait {
	return &BuildWait{
		Client:      client,
		Scheme:      scheme,
		RegistryURL: registryURL,
	}
}

// Do reconciles the build wait stage and returns the next stage that will need to be
// reconciled.  It checks if the build job has completed and if the image is available
// in the registry.
func (b *BuildWait) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx)
	logger.V(3).Info("waiting for build job to complete")

	// TODO: Check for timeout and fail.

	// There's a case where the job has already been removed.  Right now I'm sending it back to
	// the build stage, but if the image is available in the registry, we should just move on to
	// the deploy stage.

	job := GetEnvironmentJob(env, b.Scheme)
	err := b.Get(ctx, client.ObjectKeyFromObject(&job), &job)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			if b.isImageAvailable(ctx, env) {
				logger.Info("image already available", "etag", env.Spec.Revision)
				return v1beta1.EnvironmentDeployingRevision, nil
			} else {
				logger.Info("job not found, recreating")
				return v1beta1.EnvironmentCreateBuildJob, nil
			}
		}

		logger.Error(err, "unable to get job")
		return v1beta1.EnvironmentWaitingForBuildJobToComplete, err
	}

	// This provides the opportunity to check the wait status here.

	// If we are active and no completion timestamp and no conditions, requeue.

	// We care about success to move on.  There's a successful field, but I think
	// we want to check for the completion timestamp as we might want to have
	// parallel pods running in the future.

	// If failed and conditions, check the conditions for backofflimitexceeded.

	// If we have some pods that are failing, then we go into a failing state
	// and requeue. (also if not completion timestamp)

	if job.Status.Active > 0 {
		if job.Status.Failed > 0 {
			logger.V(5).Info("build job is failing")
			return v1beta1.EnvironmentStageBuildFailing, nil
		}

		if job.Status.Ready != nil && *job.Status.Ready == 0 {
			logger.V(5).Info("job is active but not ready")
			return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
		}

		if job.Status.Ready != nil && *job.Status.Ready > 0 && job.Status.CompletionTime == nil {
			logger.V(5).Info("job is active and ready")
			return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
		}
	} else {
		if job.Status.CompletionTime != nil {
			logger.V(3).Info("build completed successfully")
			return v1beta1.EnvironmentDeployingRevision, nil
		}

		if len(job.Status.Conditions) > 0 {
			next := v1beta1.EnvironmentBuildJobFailed
			logger.Error(err, "build failed", "conditions", job.Status.Conditions, "next", next)
			return next, errors.New("build failed")
		}

		// The job has more than likely failed, requeue and pick up the failure
		// status the next time through.
		return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
	}

	// if job.Status.Succeeded <= 0 && job.Status.Failed == 0 {
	// 	logger.Info("job not yet completed")
	// 	return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
	// } else if job.Status.Failed > 0 {
	// 	logger.Info("job failed")
	// 	// TODO: Return the failure reason.
	// 	return v1beta1.EnvironmentBuildJobFailed, errors.New("build job failed")
	// }

	// This is probably redundant and since we are controlling the build job we should
	// be pushing the image out there anyway.  Maybe just let the deployment fail with
	// the image errors.
	if !b.isImageAvailable(ctx, env) {
		logger.Info("image not available", "etag", env.Spec.Revision)
		return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
	}

	return v1beta1.EnvironmentDeployingRevision, nil
}

// isImageAvailable checks if the image is available in the registry.
func (b *BuildWait) isImageAvailable(ctx context.Context, env *v1beta1.Environment) bool {
	logger := log.FromContext(ctx)
	c := registry.NewClient(registry.NewHTTPClient())

	ok, err := c.HasTag(b.RegistryURL, env.GetName(), env.GetRevision())
	if err != nil {
		logger.Error(err, "unable to check image")
		return false
	}

	return ok
}

var _ Reconciler = &BuildWait{}
