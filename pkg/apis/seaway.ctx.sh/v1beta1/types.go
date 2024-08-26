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

// +kubebuilder:docs-gen:collapse=Apache License

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvironmentPort struct {
	// Name is a human readable name for the port.
	// +required
	Name string `json:"name" yaml:"name"`
	// Port is an integer representing the port number.
	// +required
	Port int32 `json:"port" yaml:"port"`
	// Protocol is the protocol for the port.  By default this is set to TCP.
	// +optional
	Protocol corev1.Protocol `json:"protocol" yaml:"protocol"`
}

type EnvironmentS3Spec struct {
	// Bucket is the name of the S3 bucket where the build context is stored.
	// +optional
	Bucket *string `json:"bucket"`
	// Prefix is the path prefix within the bucket where the build context is stored.
	// +optional
	Prefix *string `json:"prefix"`
	// Region is the AWS region where the S3 bucket is located.  This is required by
	// the minio client will always be set even for non-AWS S3 compatible services.
	// +optional
	Region *string `json:"region"`
	// Engpoint is the URL for the S3 service.
	// +optional
	Endpoint *string `json:"endpoint"`
	// ForcePathStyle represents whether or not the bucket name is a part of the
	// hostname.  This should be set for true for non-AWS S3 compatible services.
	// +optional
	ForcePathStyle *bool `json:"forcePathStyle"`
	// Credentials is a reference to a secret containing the storage credentials.
	// The secret should contain the following: AWS_ACCESS_KEY_ID and
	// AWS_SECRET_ACCESS_KEY.  Even with other providers, we stil use these environment
	// variables.
	// TODO: Right now this is a bit clunky.  I think we can do better and allow minio
	// and gcs specific credential variables.
	// +optional
	Credentials *corev1.LocalObjectReference `json:"credentials"`
	// LocalPort is used when running the sync client on a locally hosted cluster utilizing
	// k3d or kind and the s3 service for the client sync is exposed through port forwarding
	// or via an ingress.
	// LocalPort represents a port on the local machine that is forwarded to the S3 service.
	// +optional
	LocalPort *int32 `json:"localPort" yaml:"localPort"`
}

type EnvironmentBuildSpec struct {
	// Image is the build image ot use for the build job.  By default we use kaniko, but
	// this can be overridden.  However, we don't have a way to override the build command
	// at this point.
	// TODO: Override for the build command that is used.
	// +optional
	Image *string `json:"image"`
	// Platform is the platform to build the image for.  This is optional and will default
	// to the information exposed by go's runtime package.
	// +optional
	Platform *string `json:"platform"`
	// Dockerfile is the relative path inside the build context to the Dockerfile to use
	// for the build.
	// +optional
	Dockerfile *string `json:"dockerfile"`
	// Include is a list of files to include in the build context.  This is used to filter
	// out files that are not needed for the build.  They take the form of a regular expression
	// and are appended to the default includes.
	// +optional
	Include []string `json:"include"`
	// Exclude is a list of files to exclude from the build context.  This is used to filter
	// out files that are not needed for the build.  They take the form of a regular expression
	// and are appended to the default excludes.  Excludes are processed after includes so
	// if there are files in included directories that match the exclude pattern they will be
	// excluded.
	// +optional
	Exclude []string `json:"exclude"`
}

type EnvironmentVars struct {
	// Env is a list of environment variables to set in the app's container.  The environment
	// variables set here will also be used as substitution variables when the dependencies
	// are processed.
	// TODO: add the variable substitution to the dependency processing.
	// +optional
	// +nullable
	Env []corev1.EnvVar `json:"env"`
	// EnvFrom is a list of sources to populate the environment variables in the app's container.
	// +optional
	// +nullable
	EnvFrom []corev1.EnvFromSource `json:"envFrom"`
}

type EnvironmentSource struct {
	// S3 is the source for the build context.  In the future we will add other sources.
	// +optional
	S3 EnvironmentS3Spec `json:"s3"`
	// TODO: Add github as a source.
}

// EnvironmentResources is a map of corev1.ResourceName used to simplify the manifest.
// Originally I was just using the corev1.ResourceRequirements type, but it was a bit
// clunky in a manifest that you'd expect to be managed extensively by a human.
type EnvironmentResources map[corev1.ResourceName]resource.Quantity

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	// Args is a list of arguments that will be used for the deplyed application.
	// +optional
	// +nullable
	Args []string `json:"args"`
	// Build is the build spec for the environment.
	// +optional
	Build *EnvironmentBuildSpec `json:"build"`
	// Command is the command that will be used to start the deployed application.
	// +required
	Command []string `json:"command"`
	// Lifecycle is the lifecycle spec for the deployed application.
	// +optional
	// +nullable
	Lifecycle *corev1.Lifecycle `json:"lifecycle"`
	// LivenessProbe is the liveness probe for the deployed application.
	// +optional
	// +nullable
	LivenessProbe *corev1.Probe `json:"livenessProbe"`
	// Ports is a list of ports that the deployed application will listen on.  If
	// ports is not empty, the container will be created with the ports specified
	// and a service will be created to expose the ports.
	// TODO: Update the controller to create the service.
	// +optional
	// +nullable
	Ports []EnvironmentPort `json:"ports"`
	// ReadinessProbe is the readiness probe for the deployed application.
	// +optional
	// +nullable
	ReadinessProbe *corev1.Probe `json:"readinessProbe"`
	// Replicas is the number of replicas that should be deployed for the application.
	// +optional
	Replicas *int32 `json:"replicas"`
	// Resources is the resource requirements for the deployed application.
	// +optional
	// +nullable
	Resources EnvironmentResources `json:"resources"`
	// Revision is the revision of the environment.  This is used to track the revision
	// and is set by the client when the sync command is run.
	// +required
	Revision string `json:"revision"`
	// SecurityContext is the security context for the deployed application.
	// +optional
	// +nullable
	SecurityContext *corev1.SecurityContext `json:"securityContext"`
	// Source is the source for the build context.
	// +optional
	Source *EnvironmentSource `json:"source"`
	// StartupProbe is the startup probe for the deployed application.
	// +optional
	// +nullable
	StartupProbe *corev1.Probe `json:"startupProbe"`
	// Vars is a list of environment variables to set in the app's container.
	// It contains both corev1.EnvVar and corev1.EnvFromSource types.
	// +optional
	// +nullable
	Vars EnvironmentVars `json:"vars"`
	// WorkingDir is the working directory for the deployed application.
	// +optional
	// +nullable
	WorkingDir string `json:"workingDir"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=env,singular=environment
// +kubebuilder:printcolumn:name="Stage",type="string",JSONPath=".status.stage"
// +kubebuilder:printcolumn:name="Last Updated",type="date",JSONPath=".status.lastUpdated"
// +kubebuilder:printcolumn:name="Expected Revision",type="string",JSONPath=".status.expectedRevision",priority=1
// +kubebuilder:printcolumn:name="Deployed Revision",type="string",JSONPath=".status.deployedRevision",priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// EnvironmentStageis a string representation of the reconciliation stage
// of the environment.
type EnvironmentStage string

const (
	EnvironmentCreateBuildJob                 EnvironmentStage = "Creating build job"
	EnvironmentWaitingForBuildJobToComplete   EnvironmentStage = "Waiting for build job to complete"
	EnvironmentBuildJobFailed                 EnvironmentStage = "Build job failed"
	EnvironmentDeployingRevision              EnvironmentStage = "Deploying revision"
	EnvironmentWaitingForDeploymentToComplete EnvironmentStage = "Waiting for deployment to complete"
	EnvironmentDeploymentFailed               EnvironmentStage = "Deployment failed"
	EnvironmentRevisionDeployed               EnvironmentStage = "Revision deployed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvironmentStatus struct {
	// +optional
	Stage EnvironmentStage `json:"stage,omitempty"`
	// +optional
	ExpectedRevision string `json:"expectedRevision,omitempty"`
	// +optional
	DeployedRevision string `json:"deployedRevision,omitempty"`
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

// ManifestEnvironmentSpec is a spec for an environment in the manifest and
// is used by the client.
type ManifestEnvironmentSpec struct {
	Name            string `yaml:"name"`
	Namespace       string `yaml:"namespace"`
	EnvironmentSpec `yaml:",inline"`
}

// Manifest is the top level manifest definition for the client.
type Manifest struct {
	Name         string                    `yaml:"name"`
	Version      string                    `yaml:"version"`
	Description  string                    `yaml:"description"`
	Environments []ManifestEnvironmentSpec `yaml:"environments"`
}
