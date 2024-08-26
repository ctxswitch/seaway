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
	"net/url"
	"strconv"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Build is a stage that reconciles a new build job for the incoming revision.
type Build struct {
	Scheme      *runtime.Scheme
	RegistryURL string
	client.Client
}

// NewBuild creates a new Build stage.
func NewBuild(client client.Client, scheme *runtime.Scheme, registryURL string) *Build {
	return &Build{
		Client:      client,
		Scheme:      scheme,
		RegistryURL: registryURL,
	}
}

// Do reconciles the build stage and returns the next stage that will need to be
// reconciled.  It creates a new build job based on the environment.
func (b *Build) Do(ctx context.Context, env *v1beta1.Environment, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	logger := log.FromContext(ctx).WithValues("revision", env.Spec.Revision)
	logger.Info("building revision")

	status.ExpectedRevision = env.Spec.Revision

	// TODO: If the image is already present in the registry, skip the build.

	job := GetEnvironmentJob(env, b.Scheme)
	err := b.Get(ctx, client.ObjectKeyFromObject(&job), &job)
	if client.IgnoreNotFound(err) != nil {
		// The job already exists so we can't do anything with it.  Check the status and delete
		// it if it completed successfully.
		if job.Status.Succeeded > 0 {
			logger.Info("job already successfully completed, skipping")
			return v1beta1.EnvironmentDeployingRevision, nil
		} else {
			logger.Info("job already exists and has failed, skipping")
			return v1beta1.EnvironmentBuildJobFailed, nil
		}
	}

	logger.Info("creating build job")

	_, err = controllerutil.CreateOrUpdate(ctx, b.Client, &job, func() error {
		return b.buildJob(&job, env)
	})
	if err != nil {
		logger.Error(err, "unable to create or update job")
		return v1beta1.EnvironmentBuildJobFailed, err
	}

	logger.Info("job created")
	return v1beta1.EnvironmentWaitingForBuildJobToComplete, nil
}

func (b *Build) buildJob(job *batchv1.Job, env *v1beta1.Environment) error {
	job.Spec.TTLSecondsAfterFinished = ptr.To(int32(600))
	job.Spec.ActiveDeadlineSeconds = ptr.To(int64(500))
	job.Spec.BackoffLimit = ptr.To(int32(1))
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	job.Spec.PodFailurePolicy = &batchv1.PodFailurePolicy{
		Rules: []batchv1.PodFailurePolicyRule{
			{
				Action: batchv1.PodFailurePolicyActionIgnore,
				OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{
					{
						Type: corev1.DisruptionTarget,
					},
				},
			},
		},
	}

	reg, err := url.Parse(b.RegistryURL)
	if err != nil {
		return err
	}

	job.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "builder",
			Image: *env.Spec.Build.Image,
			Args: []string{
				fmt.Sprintf("--dockerfile=%s", *env.Spec.Build.Dockerfile),
				fmt.Sprintf("--context=s3://%s/%s", env.GetBucket(), env.GetKey()),
				fmt.Sprintf("--destination=%s/%s:%s", reg.Host, env.GetName(), env.GetRevision()),
				"--cache=true",
				fmt.Sprintf("--cache-repo=%s/build-cache", reg.Host),
				fmt.Sprintf("--custom-platform=%s", *env.Spec.Build.Platform),
				// Allow secure as well.
				"--insecure",
				"--insecure-pull",
				"--verbosity=trace",
			},
			EnvFrom: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: *env.Spec.Source.S3.Credentials,
					},
				},
			},
			Env: []corev1.EnvVar{
				{
					Name:  "AWS_REGION",
					Value: *env.Spec.Source.S3.Region,
				},
				{
					Name: "S3_ENDPOINT",
					// Need to add the protocol...  Either force it and strip it when setting
					// up the client or add it here.
					Value: "http://" + *env.Spec.Source.S3.Endpoint,
				},
				{
					Name:  "S3_FORCE_PATH_STYLE",
					Value: strconv.FormatBool(*env.Spec.Source.S3.ForcePathStyle),
				},
			},
		},
	}
	return nil
}

var _ Reconciler = &Build{}
