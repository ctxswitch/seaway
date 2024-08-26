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

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type ResourceInterface interface {
	Get(context.Context, interface{}, metav1.GetOptions) error
	Create(context.Context, interface{}, metav1.CreateOptions) error
	Update(context.Context, interface{}, metav1.UpdateOptions) error
}

type Object interface {
	metav1.Object
	runtime.Object
}

type ObjectKey types.NamespacedName

type OperationResult string

const (
	OperationResultCreated OperationResult = "created"
	OperationResultUpdated OperationResult = "updated"
	OperationResultNone    OperationResult = "none"
)

type MutateFn func() error
