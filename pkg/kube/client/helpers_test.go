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
