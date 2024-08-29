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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Deploy is a stage that deploys a new revision into the development environment.
type Deploy struct {
	NodePort int32
	Scheme   *runtime.Scheme
	client.Client
}

// NewDeploy creates a new Deploy stage.
func NewDeploy(client client.Client, scheme *runtime.Scheme, nodePort int32) *Deploy {
	return &Deploy{
		Client:   client,
		Scheme:   scheme,
		NodePort: nodePort,
	}
}

// Do reconciles the deploy stage and returns the next stage that will need to be
// reconciled.  It creates or updates a revision deployment based on the environment.
func (d *Deploy) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx)

	if env.Spec.Networking != nil {
		networking := env.Spec.Networking
		svc := GetEnvironmentService(env, d.Scheme)
		if networking.Ports != nil && len(networking.Ports) > 0 {
			_, err := controllerutil.CreateOrUpdate(ctx, d.Client, &svc, func() error {
				return buildService(&svc, env)
			})
			if err != nil {
				logger.Error(err, "unable to create or update service")
				return v1beta1.EnvironmentDeploymentFailed, err
			}
		}

		if networking.Ports != nil && len(networking.Ports) > 0 && networking.Ingress.Enabled {
			ing := GetEnvironmentIngress(env, d.Scheme)
			_, err := controllerutil.CreateOrUpdate(ctx, d.Client, &ing, func() error {
				return buildIngress(&ing, env)
			})
			if err != nil {
				logger.Error(err, "unable to create or update ingress")
				return v1beta1.EnvironmentDeploymentFailed, err
			}
		}
	}

	deploy := GetEnvironmentDeployment(env, d.Scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, d.Client, &deploy, func() error {
		return d.buildDeployment(&deploy, env)
	})
	if err != nil {
		logger.Error(err, "unable to create or update deployment")
		return v1beta1.EnvironmentDeploymentFailed, err
	}

	return v1beta1.EnvironmentWaitingForDeploymentToComplete, nil
}

// buildDeployment builds the deployment for the environment.
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

// buildService builds the service for the environment.
func buildService(svc *corev1.Service, env *v1beta1.Environment) error {
	ports := make([]corev1.ServicePort, 0, len(env.Spec.Networking.Ports))
	for _, port := range env.Spec.Networking.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.Port),
		})
	}

	svc.Spec.Ports = ports
	svc.Spec.Selector = map[string]string{
		"app": env.GetName(),
	}

	return nil
}

// buildIngress builds the ingress for the environment.
func buildIngress(ing *networkingv1.Ingress, env *v1beta1.Environment) error {
	ing.Spec.DefaultBackend = &networkingv1.IngressBackend{
		Service: &networkingv1.IngressServiceBackend{
			Name: env.GetName(),
			Port: networkingv1.ServiceBackendPort{
				// TODO: Right now just take the first port. I need to allow building out for
				// multiple ports in the future, but for a quick and dirty implementation,
				// this will work for now.  Just doc it and move on.
				Number: env.Spec.Networking.Ports[0].Port,
			},
		},
	}

	for a := range env.Spec.Networking.Ingress.Annotations {
		ing.Annotations[a] = env.Spec.Networking.Ingress.Annotations[a]
	}

	ing.Spec.TLS = env.Spec.Networking.Ingress.TLS

	return nil
}

var _ Reconciler = &Deploy{}
