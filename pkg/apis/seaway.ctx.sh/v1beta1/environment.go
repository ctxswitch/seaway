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

package v1beta1

import "fmt"

func (e *Environment) GetKey() string {
	return fmt.Sprintf("%s/%s-%s.tar.gz", *e.Spec.Source.S3.Prefix, e.GetName(), e.GetNamespace())
}

func (e *Environment) GetBucket() string {
	return *e.Spec.Source.S3.Bucket
}

func (e *Environment) GetRevision() string {
	return e.Spec.Revision
}

func (e *Environment) HasFailed() bool {
	return e.Status.Stage == EnvironmentBuildJobFailed || e.Status.Stage == EnvironmentDeploymentFailed
}

func (e *Environment) IsDeployed() bool {
	return e.Status.DeployedRevision == e.Spec.Revision
}

func (e *Environment) HasDeviated() bool {
	return e.Status.ExpectedRevision != "" && e.Status.ExpectedRevision != e.Spec.Revision
}
