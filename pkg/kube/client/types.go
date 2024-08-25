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
