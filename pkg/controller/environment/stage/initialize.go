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
)

type Initialize struct {
	Scheme *runtime.Scheme
	client.Client
}

// NewInitialize returns a new initialize stage.
func NewInitialize(client client.Client, scheme *runtime.Scheme) *Initialize {
	return &Initialize{
		Client: client,
		Scheme: scheme,
	}
}

// Do initializes the environment and prepares it for processing.  Initialize also provides
// a consistent entrypoint into the processing workflow.
func (b *Initialize) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	// TODO: Check to see if the expected revison already exists in the registry and if it does, then
	// short circuit the process.  This is probably an edge case if somehow the processing is interrupted
	// before the currentRevision is updated.
	status.ExpectedRevision = env.Spec.Revision
	return v1beta1.EnvironmentCheckBuildJob, nil
}
