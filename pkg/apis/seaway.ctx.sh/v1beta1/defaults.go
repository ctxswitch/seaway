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
	if obj.Replicas == nil {
		obj.Replicas = new(int32)
		*obj.Replicas = 1
	}

	defaultEnvironmentVars(&obj.Vars)
	defaultEnvironmentSource(obj.Source)
	defaultEnvironmentBuild(obj.Build)
}

func defaultEnvironmentVars(obj *EnvironmentVars) {
	if obj == nil {
		obj = new(EnvironmentVars)
	}

	if obj.Env == nil {
		obj.Env = []corev1.EnvVar{}
	}

	if obj.EnvFrom == nil {
		obj.EnvFrom = []corev1.EnvFromSource{}
	}
}

func defaultEnvironmentSource(obj *EnvironmentSource) {
	if obj == nil {
		obj = new(EnvironmentSource)
	}

	defaultEnvironmentS3Spec(&obj.S3)
}

func defaultEnvironmentBuild(obj *EnvironmentBuildSpec) {
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

	defaultEnvironmentVars(&obj.Vars)
}

func defaultEnvironmentS3Spec(obj *EnvironmentS3Spec) {
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

	// TODO: don't default credentials.  Allow access without them for cases of
	// kiam or other SA based access.
	if obj.Credentials == nil {
		obj.Credentials = new(corev1.LocalObjectReference)
		obj.Credentials.Name = DefaultCredentialsSecret
	}

	if obj.Prefix == nil {
		obj.Prefix = new(string)
		*obj.Prefix = DefaultPrefix
	}
}
