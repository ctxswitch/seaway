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
