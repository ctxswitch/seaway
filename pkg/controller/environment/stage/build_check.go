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

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BuildCheck struct {
	Scheme *runtime.Scheme
	client.Client
}

// NewBuildCheck returns a new build check stage
func NewBuildCheck(client client.Client, scheme *runtime.Scheme) *BuildCheck {
	return &BuildCheck{
		Client: client,
		Scheme: scheme,
	}
}

// Do checks for an existing build job and deletes it.
func (b *BuildCheck) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	job := GetEnvironmentJob(env, b.Scheme)
	logger := log.FromContext(ctx).WithValues("name", job.GetName())

	if err := b.Get(ctx, client.ObjectKeyFromObject(&job), &job, &client.GetOptions{}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to get job for check")
			return v1beta1.EnvironmentBuildJobFailed, err
		}
		logger.V(3).Info("job was not found, skipping deletion")
		return v1beta1.EnvironmentCreateBuildJob, nil
	}

	if !job.DeletionTimestamp.IsZero() {
		logger.V(4).Info("job is pending deletion")
		return v1beta1.EnvironmentCheckBuildJob, nil
	}

	if err := b.Delete(ctx, &job, &client.DeleteOptions{
		// If the propegation policy is not set, the pods will not be deleted
		// when the job is deleted.  For some reason it defaults to Orphan, which
		// seems like an odd default to me.
		PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
	}); err != nil {
		// We shouldn't actually hit this since we are checking above, but just in case it's here.
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "job deletion failed")
			return v1beta1.EnvironmentBuildJobFailed, err
		}
	}

	return v1beta1.EnvironmentDeletingBuildJob, nil
}
