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

package stage

import (
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GetEnvironmentJob returns a new Job object for the given environment.
func GetEnvironmentJob(env *v1beta1.Environment, scheme *runtime.Scheme) batchv1.Job {
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      env.GetName() + "-build",
			Namespace: env.GetNamespace(),
		},
	}

	controllerutil.SetControllerReference(env, &job, scheme) //nolint:errcheck

	return job
}

// GetEnvironmentDeployment returns a new Deployment object for the given environment.
func GetEnvironmentDeployment(env *v1beta1.Environment, scheme *runtime.Scheme) appsv1.Deployment {
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      env.GetName(),
			Namespace: env.GetNamespace(),
		},
	}

	controllerutil.SetControllerReference(env, &deploy, scheme) //nolint:errcheck

	return deploy
}
