package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go

// SetupWebhookWithManager adds webhook for Registry.
// func (r *Registry) SetupWebhookWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewWebhookManagedBy(mgr).
// 		For(r).
// 		Complete()
// }

// // Validate validates the Job.
// func (r *Registry) Validate() (admission.Warnings, error) {
// 	warnings := make([]string, 0)

// 	return warnings, nil
// }

// // ValidateCreate implements webhook Validator for the Registry.
// func (r *Registry) ValidateCreate() (admission.Warnings, error) {
// 	return r.Validate()
// }

// // ValidateUpdate implements webhook Validator for the Registry.
// func (r *Registry) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
// 	return r.Validate()
// }

// // ValidateDelete implements webhook Validator for the Registry.
// func (r *Registry) ValidateDelete() (admission.Warnings, error) {
// 	return nil, nil
// }

// func (r *Registry) Default() {
// 	Defaulted(r)
// }

// +kubebuilder:webhook:verbs=create;update,path=/mutate-seaway-ctx-sh-v1beta1-environment,mutating=true,failurePolicy=fail,groups=seaway.ctx.sh,resources=environments,versions=v1beta1,name=menvironment.seaway.ctx.sh,admissionReviewVersions=v1,sideEffects=none
// +kubebuilder:webhook:verbs=create;update,path=/validate-seaway-ctx-sh-v1beta1-environment,mutating=false,failurePolicy=fail,groups=seaway.ctx.sh,resources=environments,versions=v1beta1,name=venvironment.seaway.ctx.sh,admissionReviewVersions=v1,sideEffects=none

// SetupWebhookWithManager adds webhook for Watch.
func (e *Environment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(e).
		Complete()
}

// Validate validates the Job.
func (e *Environment) Validate() (admission.Warnings, error) {
	warnings := make([]string, 0)

	return warnings, nil
}

// ValidateCreate implements webhook Validator for the Watch.
func (e *Environment) ValidateCreate() (admission.Warnings, error) {
	return e.Validate()
}

// ValidateUpdate implements webhook Validator for the Watch.
func (e *Environment) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return e.Validate()
}

// ValidateDelete implements webhook Validator for the Watch.
func (e *Environment) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (e *Environment) Default() {
	Defaulted(e)
}
