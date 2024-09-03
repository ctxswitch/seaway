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

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Go

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

	// TODO: validate hostnames for ingress TLS
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
