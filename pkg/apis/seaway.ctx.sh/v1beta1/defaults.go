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
	obj.Source = defaultEnvironmentSource(obj.Source)
	obj.Build = defaultEnvironmentBuild(obj.Build)
	obj.Networking = defaultEnvironmentNetworking(obj.Networking)
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

func defaultEnvironmentSource(obj *EnvironmentSource) *EnvironmentSource {
	if obj == nil {
		obj = new(EnvironmentSource)
	}

	obj.S3 = defaultEnvironmentS3Spec(obj.S3)

	return obj
}

func defaultEnvironmentBuild(obj *EnvironmentBuildSpec) *EnvironmentBuildSpec {
	if obj == nil {
		obj = new(EnvironmentBuildSpec)
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

func defaultEnvironmentS3Spec(obj *EnvironmentS3Spec) *EnvironmentS3Spec {
	if obj == nil {
		obj = new(EnvironmentS3Spec)
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

func defaultEnvironmentNetworking(obj *EnvironmentNetworking) *EnvironmentNetworking {
	if obj == nil {
		obj = new(EnvironmentNetworking)
	}

	if obj.Ports == nil {
		obj.Ports = []EnvironmentPort{}
	}

	for _, port := range obj.Ports {
		defaultEnvironmentPort(&port)
	}

	obj.Ingress = defaultEnvironmentIngress(obj.Ingress)

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

func defaultEnvironmentIngress(obj *EnvironmentIngress) *EnvironmentIngress {
	if obj == nil {
		obj = new(EnvironmentIngress)
	}

	if obj.Annotations == nil {
		obj.Annotations = map[string]string{}
	}

	return obj
}
