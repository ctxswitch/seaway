package v1beta1

import (
	"runtime"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Defaulted(obj client.Object) {
	switch v := obj.(type) {
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
		*obj.Image = "gcr.io/kaniko-project/executor:latest"
	}

	if obj.Dockerfile == nil {
		obj.Dockerfile = new(string)
		*obj.Dockerfile = "Dockerfile"
	}

	if obj.Platform == nil {
		obj.Platform = new(string)
		*obj.Platform = runtime.GOOS + "/" + runtime.GOARCH
	}

	if obj.Include == nil {
		obj.Include = []string{}
	}

	if obj.Exclude == nil {
		obj.Exclude = []string{}
	}
}

func defaultEnvironmentS3Spec(obj *EnvironmentS3Spec) {
	if obj == nil {
		obj = new(EnvironmentS3Spec)
	}

	if obj.Bucket == nil {
		obj.Bucket = new(string)
		*obj.Bucket = "seaway"
	}

	if obj.Region == nil {
		obj.Region = new(string)
		*obj.Region = "us-east-1"
	}

	if obj.Endpoint == nil {
		obj.Endpoint = new(string)
		*obj.Endpoint = "http://localhost:80"
	}

	if obj.ForcePathStyle == nil {
		obj.ForcePathStyle = new(bool)
		*obj.ForcePathStyle = true
	}

	// TODO: don't default credentials.  Allow access without them for cases of
	// kiam or other SA based access.
	if obj.Credentials == nil {
		obj.Credentials = new(corev1.LocalObjectReference)
		obj.Credentials.Name = "seaway-credentials"
	}

	if obj.Prefix == nil {
		obj.Prefix = new(string)
		*obj.Prefix = "contexts"
	}
}
