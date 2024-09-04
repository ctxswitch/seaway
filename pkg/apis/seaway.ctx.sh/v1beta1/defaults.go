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
	"runtime"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultReplicas          = 1
	DefaultBucket            = "seaway"
	DefaultRegion            = "us-east-1"
	DefaultEndpoint          = "http://localhost:80"
	DefaultForcePathStyle    = true
	DefaultPrefix            = "artifacts"
	DefaultBuildImage        = "gcr.io/kaniko-project/executor:latest"
	DefaultDockerfile        = "Dockerfile"
	DefaultPlatform          = runtime.GOOS + "/" + runtime.GOARCH
	DefaultCredentialsSecret = "seaway-s3-credentials" //nolint:gosec
)

func Defaulted(obj client.Object) {
	switch v := obj.(type) { //nolint:gocritic
	case *Environment:
		defaultEnvironment(v)
	}
}

func defaultEnvironment(obj *Environment) {
	defaultEnvironmentSpec(&obj.Spec)
}

func defaultEnvironmentSpec(obj *EnvironmentSpec) {
	if obj == nil {
		obj = new(EnvironmentSpec)
	}

	if obj.Args == nil {
		obj.Args = []string{}
	}

	if obj.Resources == nil {
		obj.Resources = EnvironmentResources{}
	}

	if obj.Replicas == nil {
		obj.Replicas = new(int32)
		*obj.Replicas = DefaultReplicas
	}

	obj.Vars = defaultEnvironmentVars(obj.Vars)
	obj.Store = defaultEnvironmentStore(obj.Store)
	obj.Build = defaultEnvironmentBuild(obj.Build)
	obj.Network = defaultEnvironmentNetwork(obj.Network)
}

func defaultEnvironmentVars(obj *EnvironmentVars) *EnvironmentVars {
	if obj == nil {
		obj = new(EnvironmentVars)
	}

	if obj.Env == nil {
		obj.Env = []corev1.EnvVar{}
	}

	if obj.EnvFrom == nil {
		obj.EnvFrom = []corev1.EnvFromSource{}
	}

	return obj
}

func defaultEnvironmentBuild(obj *EnvironmentBuild) *EnvironmentBuild {
	if obj == nil {
		obj = new(EnvironmentBuild)
	}

	// RegistryRef is required.
	if obj.Image == nil {
		obj.Image = new(string)
		*obj.Image = DefaultBuildImage
	}

	if obj.Dockerfile == nil {
		obj.Dockerfile = new(string)
		*obj.Dockerfile = DefaultDockerfile
	}

	if obj.Platform == nil {
		obj.Platform = new(string)
		*obj.Platform = DefaultPlatform
	}

	if obj.Include == nil {
		obj.Include = []string{}
	}

	if obj.Exclude == nil {
		obj.Exclude = []string{}
	}

	return obj
}

func defaultEnvironmentStore(obj *EnvironmentStore) *EnvironmentStore {
	if obj == nil {
		obj = new(EnvironmentStore)
	}

	if obj.Bucket == nil {
		obj.Bucket = new(string)
		*obj.Bucket = DefaultBucket
	}

	if obj.Region == nil {
		obj.Region = new(string)
		*obj.Region = DefaultRegion
	}

	if obj.Endpoint == nil {
		obj.Endpoint = new(string)
		*obj.Endpoint = DefaultEndpoint
	}

	if obj.ForcePathStyle == nil {
		obj.ForcePathStyle = new(bool)
		*obj.ForcePathStyle = DefaultForcePathStyle
	}

	if obj.Prefix == nil {
		obj.Prefix = new(string)
		*obj.Prefix = DefaultPrefix
	}

	return obj
}

func defaultEnvironmentNetwork(obj *EnvironmentNetwork) *EnvironmentNetwork {
	if obj == nil {
		obj = new(EnvironmentNetwork)
	}

	obj.Service = defaultEnvironmentService(obj.Service)

	if obj.Service.Enabled {
		obj.Ingress = defaultEnvironmentIngress(obj.Ingress, obj.Service.Ports[0].Port)
	}

	return obj
}

func defaultEnvironmentPort(obj *EnvironmentPort) *EnvironmentPort {
	if obj == nil {
		obj = new(EnvironmentPort)
	}

	if obj.Protocol == "" {
		obj.Protocol = corev1.ProtocolTCP
	}

	return obj
}

func defaultEnvironmentService(obj *EnvironmentService) *EnvironmentService {
	if obj == nil {
		obj = new(EnvironmentService)
	}

	if !obj.Enabled {
		return obj
	}

	if obj.Annotations == nil {
		obj.Annotations = map[string]string{}
	}

	if obj.Type == "" {
		obj.Type = corev1.ServiceTypeClusterIP
	}

	if len(obj.Ports) == 0 {
		obj.Ports = []EnvironmentPort{
			{
				Name:     "http",
				Port:     9000,
				Protocol: corev1.ProtocolTCP,
			},
		}
	} else {
		for _, port := range obj.Ports {
			defaultEnvironmentPort(&port)
		}
	}

	return obj
}

func defaultEnvironmentIngress(obj *EnvironmentIngress, defaultPort int32) *EnvironmentIngress {
	if obj == nil {
		obj = new(EnvironmentIngress)
	}

	if !obj.Enabled {
		return obj
	}

	if obj.Annotations == nil {
		obj.Annotations = map[string]string{}
	}

	if obj.Port == nil {
		obj.Port = new(int32)
		*obj.Port = defaultPort
	}

	return obj
}
