package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type MockObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (o *MockObject) DeepCopyObject() runtime.Object { return o }

func TestObjectKeyFromObject(t *testing.T) {
	obj := &MockObject{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
	}

	expected := ObjectKey{
		Namespace: "test",
		Name:      "test",
	}

	got := ObjectKeyFromObject(obj)

	assert.Equal(t, expected, got)
}
