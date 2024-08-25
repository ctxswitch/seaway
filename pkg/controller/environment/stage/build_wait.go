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

type BuildWait struct {
	Scheme      *runtime.Scheme
	RegistryURL string
	client.Client
}

func NewBuildWait(client client.Client, scheme *runtime.Scheme, registryURL string) *BuildWait {
	return &BuildWait{
		Client:      client,
		Scheme:      scheme,
		RegistryURL: registryURL,
	}
}

func (b *BuildWait) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentCondition, error) {
	logger := log.FromContext(ctx).WithValues("revision", env.Spec.Revision)
	logger.Info("waiting for build job to complete")

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

	if job.Status.Succeeded <= 0 && job.Status.Failed == 0 {
		logger.Info("job not yet completed")
		return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
	} else if job.Status.Failed > 0 {
		logger.Info("job failed")
		// TODO: Return the failure reason.
		return v1beta1.EnvironmentBuildJobFailed, errors.New("build job failed")
	}

	// This is probably redundant and since we are controlling the build job we should
	// be pushing the image out there anyway.  Maybe just let the deployment fail with
	// the image errors.
	if !b.isImageAvailable(ctx, env) {
		logger.Info("image not available", "etag", env.Spec.Revision)
		return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
	}

	return v1beta1.EnvironmentDeployingRevision, nil
}

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
