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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	// NodePort is the port on each node that the service is exposed on when the
	// service type is NodePort or LoadBalancer.  Type must be set to NodePort
	// or LoadBalancer for this field to have an effect.
	// +optional
	NodePort int32 `json:"nodePort" yaml:"nodePort"`
}

type EnvironmentIngress struct {
	// Annotations is a map of annotations to apply to the ingress resource.
	// +optional
	// +nullable
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
	// Enabled is a flag to enable or disable the ingress resource.  It is disabled by default.
	// +optional
	Enabled bool `json:"enabled" yaml:"enabled"`
	// ClassName is the name of the ingress class to use for the ingress resource.
	// +optional
	// +nullable
	ClassName *string `json:"className" yaml:"className"`
	// Port is the port on the service that the ingress will route traffic to. By
	// default this will pick the first port listed in the service.
	// +optional
	// +nullable
	Port *int32 `json:"port" yaml:"port"`
	// TLS is a list of TLS configuration for the ingress resource.  The TLS configuration
	// matches that of the networking.k8s.io/v1beta1 Ingress type.
	// +optional
	// +nullable
	TLS []networkingv1.IngressTLS `json:"tls" yaml:"tls"`
}

type EnvironmentService struct {
	// Annotations is a map of annotations to apply to the service resource.
	// +optional
	// +nullable
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
	// Enabled is a flag to enable or disable the service resource.  It is disabled by
	// default.
	// +optional
	Enabled bool `json:"enabled" yaml:"enabled"`
	// ExternalName is the external reference that discovery will use as an alias for
	// the service (CNAME).
	// +optional
	// +nullable
	ExternalName *string `json:"externalName" yaml:"externalName"`
	// Ports is a list of ports that the deployed application will listen on.  If the
	// service is enabled and the ports are not set, the controller will default to
	// port 9000.
	// +optional
	// +nullable
	Ports []EnvironmentPort `json:"ports" yaml:"ports"`
	// Type is the type of service to create.  By default this is set to ClusterIP.
	// +optional
	Type corev1.ServiceType `json:"type" yaml:"type"`
}

type EnvironmentNetwork struct {
	// Ingress is the ingress configuration for the deployed application.  If enabled, the
	// controller will create an ingress resource to expose the application.
	// +optional
	// +nullable
	Ingress *EnvironmentIngress `json:"ingress" yaml:"ingress"`
	// Service is the service configuration for the deployed application.
	// +optional
	// +nullable
	Service *EnvironmentService `json:"service" yaml:"service"`
}

type EnvironmentBuild struct {
	// Args are the command arguments that will be passed to the build job.
	// +optional
	// +nullable
	Args []string `json:"args" yaml:"args"`
	// Command is the command that will be passed to the build job.
	// +optional
	// +nullable
	Command []string `json:"command" yaml:"command"`
	// Image is the build image ot use for the build job.  By default we use kaniko, but
	// this can be overridden.  However, we don't have a way to override the build command
	// at this point.
	// TODO: Override for the build command that is used.
	// +optional
	Image *string `json:"image" yaml:"image"`
	// Platform is the platform to build the image for.  This is optional and will default
	// to the information exposed by go's runtime package.
	// +optional
	Platform *string `json:"platform" yaml:"platform"`
	// Dockerfile is the relative path inside the build context to the Dockerfile to use
	// for the build.
	// +optional
	Dockerfile *string `json:"dockerfile" yaml:"dockerfile"`
	// Include is a list of files to include in the build context.  This is used to filter
	// out files that are not needed for the build.  They take the form of a regular expression
	// and are appended to the default includes.
	// +optional
	Include []string `json:"include" yaml:"include"`
	// Exclude is a list of files to exclude from the build context.  This is used to filter
	// out files that are not needed for the build.  They take the form of a regular expression
	// and are appended to the default excludes.  Excludes are processed after includes so
	// if there are files in included directories that match the exclude pattern they will be
	// excluded.
	// +optional
	Exclude []string `json:"exclude" yaml:"exclude"`
}

type EnvironmentVars struct {
	// Env is a list of environment variables to set in the app's container.  The environment
	// variables set here will also be used as substitution variables when the dependencies
	// are processed.
	// TODO: add the variable substitution to the dependency processing.
	// +optional
	// +nullable
	Env []corev1.EnvVar `json:"env" yaml:"env"`
	// EnvFrom is a list of sources to populate the environment variables in the app's container.
	// +optional
	// +nullable
	EnvFrom []corev1.EnvFromSource `json:"envFrom" yaml:"envFrom"`
}

// EnvironmentResources is a map of corev1.ResourceName used to simplify the manifest.
// Originally I was just using the corev1.ResourceRequirements type, but it was a bit
// clunky in a manifest that you'd expect to be managed extensively by a human.
type EnvironmentResources map[corev1.ResourceName]resource.Quantity

// EnvironmentSpec defines the desired state of Environment.
type EnvironmentSpec struct {
	// Args is a list of arguments that will be used for the deployed application.
	// +optional
	// +nullable
	Args []string `json:"args" yaml:"args"`
	// Build is the build spec for the environment.
	// +optional
	Build *EnvironmentBuild `json:"build" yaml:"build"`
	// Config
	// +optional
	Config string `json:"config" yaml:"config"`
	// Command is the command that will be used to start the deployed application.
	// +optional
	// +nullable
	Command []string `json:"command" yaml:"command"`
	// Endpoint is the Seaway API endpoint that the client will use to interact
	// with the environment.
	// +optional
	Endpoint *string `json:"endpoint" yaml:"endpoint"`
	// Lifecycle is the lifecycle spec for the deployed application.
	// +optional
	// +nullable
	Lifecycle *corev1.Lifecycle `json:"lifecycle" yaml:"lifecycle"`
	// LivenessProbe is the liveness probe for the deployed application.
	// +optional
	// +nullable
	LivenessProbe *corev1.Probe `json:"livenessProbe" yaml:"livenessProbe"`
	// Network contains the network configuration options for the environment.
	// +optional
	Network *EnvironmentNetwork `json:"network" yaml:"network"`
	// ReadinessProbe is the readiness probe for the deployed application.
	// +optional
	// +nullable
	ReadinessProbe *corev1.Probe `json:"readinessProbe" yaml:"readinessProbe"`
	// Replicas is the number of replicas that should be deployed for the application.
	// +optional
	Replicas *int32 `json:"replicas" yaml:"replicas"`
	// Resources is the resource requirements for the deployed application.
	// +optional
	Resources EnvironmentResources `json:"resources" yaml:"resources"`
	// Revision is the revision of the environment.  This is used to track the revision
	// and is set by the client when the sync command is run.
	// +required
	Revision string `json:"revision" yaml:"revision"`
	// SecurityContext is the security context for the deployed application.
	// +optional
	// +nullable
	SecurityContext *corev1.SecurityContext `json:"securityContext" yaml:"securityContext"`

	// StartupProbe is the startup probe for the deployed application.
	// +optional
	// +nullable
	StartupProbe *corev1.Probe `json:"startupProbe" yaml:"startupProbe"`
	// Vars is a list of environment variables to set in the app's container.
	// It contains both corev1.EnvVar and corev1.EnvFromSource types.
	// +optional
	// +nullable
	Vars *EnvironmentVars `json:"vars" yaml:"vars"`
	// WorkingDir is the working directory for the deployed application.
	// +optional
	// +nullable
	WorkingDir string `json:"workingDir" yaml:"workingDir"`
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
	EnvironmentStageInitialize        EnvironmentStage = ""
	EnvironmentStageBuildImage        EnvironmentStage = "Creating the build job"
	EnvironmentStageBuildImageWait    EnvironmentStage = "Waiting for build to complete"
	EnvironmentStageBuildImageFailing EnvironmentStage = "Build job is failing"
	EnvironmentStageBuildImageFailed  EnvironmentStage = "Build failed"
	EnvironmentStageBuildImageVerify  EnvironmentStage = "Verifying the image"
	EnvironmentStageDeploy            EnvironmentStage = "Deploying the revision"
	EnvironmentStageDeployWaiting     EnvironmentStage = "Waiting for deployment to complete"
	EnvironmentStageDeployVerify      EnvironmentStage = "Verifying the deployment"
	EnvironmentStageDeployed          EnvironmentStage = "Revision deployed"
	EnvironmentStageDeployFailed      EnvironmentStage = "Deployment failed"
	EnvironmentStageFailed            EnvironmentStage = "Server error"
)

type EnvironmentStatus struct {
	// +optional
	Stage EnvironmentStage `json:"stage,omitempty"`
	// +optional
	ExpectedRevision string `json:"expectedRevision,omitempty"`
	// +optional
	DeployedRevision string `json:"deployedRevision,omitempty"`
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

type DependencyType string

const (
	Kustomize DependencyType = "kustomize"
)

// ManifestDependency is a dependency configuration that can be applied to
// the environment.  Only kustomize is supported at this time.
type ManifestDependency struct {
	// Name is the name of the dependency.
	// +required
	Name string `yaml:"name"`
	// Type is the type of dependency.  Only kustomize is supported at this time.
	// +optional
	Type DependencyType `yaml:"type"`
	// Path is the path to the directory containing the manifests.
	// +required
	Path string `yaml:"path"`
}

// ManifestEnvironmentSpec is a spec for an environment in the manifest and
// is used by the client.
type ManifestEnvironmentSpec struct {
	Name            string               `yaml:"name"`
	Namespace       string               `yaml:"namespace"`
	Dependencies    []ManifestDependency `yaml:"dependencies"`
	EnvironmentSpec `yaml:",inline"`
}

// Manifest is the top level manifest definition for the client.
type Manifest struct {
	Name         string                    `yaml:"name"`
	Version      string                    `yaml:"version"`
	Description  string                    `yaml:"description"`
	Environments []ManifestEnvironmentSpec `yaml:"environments"`
}

// SeawayConfigStorageSpec is the storage configuration for the controller.
type SeawayConfigStorageSpec struct {
	// Bucket is the name of the bucket to use for storage.
	// +required
	Bucket string `json:"bucket"`
	// Endpoint is the endpoint for the storage service.
	// +required
	Endpoint string `json:"endpoint"`
	// Prefix is the prefix to use for the storage objects.
	// +optional
	Prefix string `json:"prefix"`
	// Region is the region to use for the storage service.
	// +required
	Region string `json:"region"`
	// Credentials is the name of the secret that contains the credentials
	// for the storage service.
	// +required
	Credentials string `json:"credentials"`
	// ForcePathStyle is a flag to force path style addressing for the storage
	// service.
	// +optional
	ForcePathStyle bool `json:"forcePathStyle"`
}

// SeawayConfigRegistrySpec is the registry configuration for the controller.
type SeawayConfigRegistrySpec struct {
	// URL is the URL for the registry service.
	// +required
	URL string `json:"url"`
	// NodePort is the node port for the registry service.
	// +required
	NodePort int32 `json:"nodePort"`
}

type SeawayConfigSpec struct {
	SeawayConfigStorageSpec  `json:"storage,omitempty"`
	SeawayConfigRegistrySpec `json:"registry,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:resource:scope=Namespaced,shortName=sc,singular=seawayconfig
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Config is the configuration that the controller will use.  It contains the
// global configurations for the controller.
type SeawayConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SeawayConfigSpec `json:"spec,omitempty"`
}

func (c *SeawayConfigStorageSpec) GetArchiveKey(name, namespace string) string {
	if c.Prefix == "" {
		return fmt.Sprintf("%s-%s.tar.gz", name, namespace)
	}

	return fmt.Sprintf("%s/%s-%s.tar.gz", c.Prefix, name, namespace)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SeawayConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeawayConfig `json:"items"`
}
