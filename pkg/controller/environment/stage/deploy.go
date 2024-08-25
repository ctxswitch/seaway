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
	"context"
	"fmt"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Deploy struct {
	NodePort int32
	Scheme   *runtime.Scheme
	client.Client
}

func NewDeploy(client client.Client, scheme *runtime.Scheme, nodePort int32) *Deploy {
	return &Deploy{
		Client:   client,
		Scheme:   scheme,
		NodePort: nodePort,
	}
}

func (d *Deploy) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentCondition, error) {
	logger := log.FromContext(ctx)

	deploy := GetEnvironmentDeployment(env, d.Scheme)

	// TODO: Get and check the deployment annotations for the revision and stop if the revision
	// is already deployed?

	_, err := controllerutil.CreateOrUpdate(ctx, d.Client, &deploy, func() error {
		return d.buildDeployment(&deploy, env)
	})
	if err != nil {
		logger.Error(err, "unable to create or update deployment")
		return v1beta1.EnvironmentDeploymentFailed, err
	}

	return v1beta1.EnvironmentWaitingForDeploymentToComplete, nil
}

func (d *Deploy) buildDeployment(deploy *appsv1.Deployment, env *v1beta1.Environment) error {
	container := corev1.Container{
		Name:           "app",
		Image:          fmt.Sprintf("localhost:%d/%s:%s", d.NodePort, env.GetName(), env.GetRevision()),
		Command:        env.Spec.Command,
		Args:           env.Spec.Args,
		WorkingDir:     env.Spec.WorkingDir,
		Ports:          env.Spec.ContainerPorts(),
		EnvFrom:        env.Spec.Vars.EnvFrom,
		Env:            env.Spec.Vars.Env,
		Resources:      env.Spec.Resources.CoreV1ResourceRequirements(),
		LivenessProbe:  env.Spec.LivenessProbe,
		ReadinessProbe: env.Spec.ReadinessProbe,
		StartupProbe:   env.Spec.StartupProbe,
		Lifecycle:      env.Spec.Lifecycle,
	}

	deploy.Spec.Replicas = env.Spec.Replicas
	// TODO: Add additional configuration.  We should also add a way to add sidecars.
	deploy.Spec.Template.Spec.Containers = []corev1.Container{container}

	if deploy.ObjectMeta.CreationTimestamp.IsZero() {
		deploy.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  env.GetName(),
				"etag": env.GetRevision(),
			},
		}
		deploy.Spec.Template.Labels = map[string]string{
			"app":  env.GetName(),
			"etag": env.GetRevision(),
		}
	}

	return nil
}

var _ Reconciler = &Deploy{}
