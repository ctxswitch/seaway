package v1beta1

// +kubebuilder:docs-gen:collapse=Apache License

import (
	seawayv1beta1 "ctx.sh/seaway/pkg/apis/seaway.ctx.sh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +kubebuilder:docs-gen:collapse=Go imports

// Version specifies the API version
const Version = "v1beta1"

// SchemaGroupVersion is the group version that wil be used to register the objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   seawayv1beta1.GroupName,
	Version: Version,
}

// Kind takes an unqualified kind and returns back a Group and qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns back a Group and qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// addKnownTypes adds a list of known types to the scheme
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Environment{},
		&EnvironmentList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
