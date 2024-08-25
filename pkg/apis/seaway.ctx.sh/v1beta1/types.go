package v1beta1

// +kubebuilder:docs-gen:collapse=Apache License

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvironmentPort struct {
	// +required
	Name string `json:"name" yaml:"name"`
	// +required
	Port int32 `json:"port" yaml:"port"`
	// +optional
	Protocol corev1.Protocol `json:"protocol" yaml:"protocol"`
}

type EnvironmentS3Spec struct {
	// +optional
	Bucket *string `json:"bucket"`
	// +optional
	Prefix *string `json:"prefix"`
	// +optional
	Region *string `json:"region"`
	// +optional
	Endpoint *string `json:"endpoint"`
	// +optional
	ForcePathStyle *bool `json:"forcePathStyle"`
	// +optional
	Credentials *corev1.LocalObjectReference `json:"credentials"`
	// LocalPort is used when running the sync client on a locally hosted cluster utilizing
	// k3d or kind and the s3 service for the client sync is exposed through port forwarding
	// or via an ingress.
	// +optional
	LocalPort *int32 `json:"localPort" yaml:"localPort"`
}

type EnvironmentBuildSpec struct {
	// +optional
	Image *string `json:"image"`
	// +optional
	Platform *string `json:"platform"`
	// +optional
	Dockerfile *string `json:"dockerfile"`
	// +optional
	Include []string `json:"include"`
	// +optional
	Exclude []string `json:"exclude"`
}

type EnvironmentVars struct {
	// +optional
	// +nullable
	Env []corev1.EnvVar `json:"env"`
	// +optional
	// +nullable
	EnvFrom []corev1.EnvFromSource `json:"envFrom"`
}

type EnvironmentSource struct {
	// +optional
	S3 EnvironmentS3Spec `json:"s3"`
}

type EnvironmentResources map[corev1.ResourceName]resource.Quantity

type EnvironmentSpec struct {
	// +optional
	// +nullable
	Args []string `json:"args"`
	// +optional
	Build *EnvironmentBuildSpec `json:"build"`
	// +required
	Command []string `json:"command"`
	// +optional
	// +nullable
	Lifecycle *corev1.Lifecycle `json:"lifecycle"`
	// +optional
	// +nullable
	LivenessProbe *corev1.Probe `json:"livenessProbe"`
	// +optional
	// +nullable
	Ports []EnvironmentPort `json:"ports"`
	// +optional
	// +nullable
	ReadinessProbe *corev1.Probe `json:"readinessProbe"`
	// +optional
	Replicas *int32 `json:"replicas"`
	// +optional
	// +nullable
	Resources EnvironmentResources `json:"resources"`
	// +required
	Revision string `json:"revision"`
	// +optional
	// +nullable
	SecurityContext *corev1.SecurityContext `json:"securityContext"`
	// +optional
	Source *EnvironmentSource `json:"source"`
	// +optional
	// +nullable
	StartupProbe *corev1.Probe `json:"startupProbe"`
	// +optional
	// +nullable
	Vars EnvironmentVars `json:"vars"`
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

type EnvironmentCondition string

const (
	EnvironmentCreateBuildJob                 EnvironmentCondition = "Creating build job"
	EnvironmentWaitingForBuildJobToComplete   EnvironmentCondition = "Waiting for build job to complete"
	EnvironmentBuildJobFailed                 EnvironmentCondition = "Build job failed"
	EnvironmentDeployingRevision              EnvironmentCondition = "Deploying revision"
	EnvironmentWaitingForDeploymentToComplete EnvironmentCondition = "Waiting for deployment to complete"
	EnvironmentDeploymentFailed               EnvironmentCondition = "Deployment failed"
	EnvironmentRevisionDeployed               EnvironmentCondition = "Revision deployed"
)

type EnvironmentStatus struct {
	// +optional
	Stage EnvironmentCondition `json:"stage,omitempty"`
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

type ManifestEnvironmentSpec struct {
	Name            string `yaml:"name"`
	Namespace       string `yaml:"namespace"`
	EnvironmentSpec `yaml:",inline"`
}

type Manifest struct {
	Name         string                    `yaml:"name"`
	Version      string                    `yaml:"version"`
	Description  string                    `yaml:"description"`
	Environments []ManifestEnvironmentSpec `yaml:"environments"`
}
