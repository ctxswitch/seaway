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

// GetKey returns the key for the archive object in the S3 bucket
func (e *Environment) GetKey() string {
	return fmt.Sprintf("%s/%s-%s.tar.gz", *e.Spec.Source.S3.Prefix, e.GetName(), e.GetNamespace())
}

// GetBucket returns the bucket where the archive object is stored
func (e *Environment) GetBucket() string {
	return *e.Spec.Source.S3.Bucket
}

// GetRevision returns the configured revision of the environment
func (e *Environment) GetRevision() string {
	return e.Spec.Revision
}

// HasFailed returns true if the environment has failed to build or deploy
func (e *Environment) HasFailed() bool {
	return e.Status.Stage == EnvironmentBuildJobFailed || e.Status.Stage == EnvironmentDeploymentFailed
}

// IsDeployed returns true if the environment has been deployed.  At the end of the
// reconciliation loop, the status of the environment is updated to reflect the
// successfully deployed revision so we can check to see if the spec revision matches
// the deployed revision in the status.
func (e *Environment) IsDeployed() bool {
	return e.Status.DeployedRevision == e.Spec.Revision
}

// HasDeviated returns true if the configured revision has deviated from the deployed
// revision.
func (e *Environment) HasDeviated() bool {
	return e.Status.ExpectedRevision != "" && e.Status.ExpectedRevision != e.Spec.Revision
}
