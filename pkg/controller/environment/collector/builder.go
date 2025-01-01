package collector

import (
	"fmt"
	"net/url"
	"strconv"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

type DesiredState struct {
	Job            *batchv1.Job
	Deployment     *appsv1.Deployment
	Service        *corev1.Service
	Ingress        *networkingv1.Ingress
	Config         *v1beta1.EnvironmentConfig
	EnvCredentials *corev1.Secret
	BuildNamespace *corev1.Namespace
}

func NewDesiredState() *DesiredState {
	return &DesiredState{
		Job:            nil,
		Deployment:     nil,
		Service:        nil,
		Ingress:        nil,
		Config:         nil,
		EnvCredentials: nil,
		BuildNamespace: nil,
	}
}

type Builder struct {
	observed *ObservedState
	scheme   *runtime.Scheme

	registry    v1beta1.EnvironmentConfigRegistrySpec
	registryURL *url.URL
	storage     v1beta1.EnvironmentConfigStorageSpec
	storageURL  *url.URL

	builderNamespace string
}

func (b *Builder) desired(d *DesiredState) error {
	env := b.observed.Env
	if env == nil {
		return nil
	}

	var err error

	b.storage = b.observed.Config.Spec.EnvironmentConfigStorageSpec
	b.storageURL, err = url.Parse(b.storage.Endpoint)
	if err != nil {
		return err
	}

	b.registry = b.observed.Config.Spec.EnvironmentConfigRegistrySpec
	b.registryURL, err = url.Parse(b.registry.URL)
	if err != nil {
		return err
	}

	d.BuildNamespace = b.buildNamespace(b.builderNamespace)
	d.EnvCredentials = b.buildEnvCredentials()
	d.Job = b.buildJob()
	d.Deployment = b.buildDeployment()

	if env.Spec.Network.Service.Enabled {
		d.Service = b.buildService()
		if env.Spec.Network.Ingress.Enabled {
			d.Ingress = b.buildIngress()
		}
	}

	return nil
}

func (b *Builder) buildNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func (b *Builder) buildEnvCredentials() *corev1.Secret {
	env := b.observed.Env
	credentials := b.observed.EnvCredentials

	if credentials == nil {
		credentials = &corev1.Secret{}
	}

	metatdata := metav1.ObjectMeta{
		Name:      env.Name + "-credentials",
		Namespace: env.Namespace,
		Annotations: mergeMap(map[string]string{
			"seaway.ctx.sh/revision": env.GetRevision(),
		}, credentials.Annotations),
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	data := b.observed.StorageCredentials.Data

	return &corev1.Secret{
		ObjectMeta: metatdata,
		Data:       data,
	}
}

func (b *Builder) buildJob() *batchv1.Job { //nolint:funlen
	env := b.observed.Env

	metatdata := metav1.ObjectMeta{
		Name:      env.Name + "-build",
		Namespace: env.Namespace,
		Annotations: map[string]string{
			"seaway.ctx.sh/revision": env.GetRevision(),
		},
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	args := env.Spec.Build.Args
	if args == nil {
		args = []string{
			fmt.Sprintf("--dockerfile=%s", *env.Spec.Build.Dockerfile),
			fmt.Sprintf("--context=s3://%s/%s", b.storage.Bucket, b.storage.GetArchiveKey(env.GetName(), env.GetNamespace())),
			fmt.Sprintf("--destination=%s/%s:%s", b.registryURL.Host, env.GetName(), env.GetRevision()),
			// TODO: toggle caching
			"--cache=true",
			fmt.Sprintf("--cache-repo=%s/build-cache", b.registryURL.Host),
			fmt.Sprintf("--custom-platform=%s", *env.Spec.Build.Platform),
			// TODO: Allow secure as well based on the registry uri parsing.
			"--insecure",
			"--insecure-pull",
			// TODO: Make this configurable
			"--verbosity=info",
		}
	}

	vars := []corev1.EnvVar{
		{
			Name:  "AWS_REGION",
			Value: b.storage.Region,
		},
		{
			Name: "S3_ENDPOINT",
			// Need to add the protocol...  Either force it and strip it when setting
			// up the client or add it here.
			Value: b.storage.Endpoint,
		},
		{
			Name:  "S3_FORCE_PATH_STYLE",
			Value: strconv.FormatBool(b.storage.ForcePathStyle),
		},
	}

	container := corev1.Container{
		Name:    "builder",
		Image:   *env.Spec.Build.Image,
		Command: env.Spec.Build.Command,
		Args:    args,
		Env:     mergeEnvVar(vars, env.Spec.Vars.Env),
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: env.Name + "-credentials",
					},
				},
			},
		},
	}

	spec := batchv1.JobSpec{
		// TODO: ttl, activedeadline and backoff should be configurable
		TTLSecondsAfterFinished: ptr.To(int32(3600)),
		ActiveDeadlineSeconds:   ptr.To(int64(600)),
		BackoffLimit:            ptr.To(int32(1)),
		PodFailurePolicy: &batchv1.PodFailurePolicy{
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
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":   env.GetName(),
					"etag":  env.GetRevision(),
					"group": "build",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers:    []corev1.Container{container},
			},
		},
	}

	return &batchv1.Job{
		ObjectMeta: metatdata,
		Spec:       spec,
	}
}

func (b *Builder) buildDeployment() *appsv1.Deployment {
	env := b.observed.Env
	deployment := b.observed.Deployment

	if deployment == nil {
		deployment = &appsv1.Deployment{}
	}

	metatdata := metav1.ObjectMeta{
		Name:      env.Name,
		Namespace: env.Namespace,
		Annotations: mergeMap(map[string]string{
			"seaway.ctx.sh/revision": env.GetRevision(),
		}, deployment.Annotations),
		Labels: mergeMap(map[string]string{
			"app":  env.GetName(),
			"etag": env.GetRevision(),
		}, deployment.Labels),
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	container := corev1.Container{
		Name:           "app",
		Image:          fmt.Sprintf("localhost:%d/%s:%s", b.registry.NodePort, env.GetName(), env.GetRevision()),
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

	spec := appsv1.DeploymentSpec{
		Replicas: env.Spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":   env.GetName(),
				"group": "application",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":   env.GetName(),
					"group": "application",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{container},
			},
		},
	}

	return &appsv1.Deployment{
		ObjectMeta: metatdata,
		Spec:       spec,
	}
}

func (b *Builder) buildService() *corev1.Service {
	env := b.observed.Env
	service := b.observed.Service

	if service == nil {
		service = &corev1.Service{}
	}

	metatdata := metav1.ObjectMeta{
		Name:        env.Name,
		Namespace:   env.Namespace,
		Annotations: service.Annotations,
		Labels:      service.Labels,
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	ports := make([]corev1.ServicePort, 0, len(env.Spec.Network.Service.Ports))
	for _, port := range env.Spec.Network.Service.Ports {
		servicePort := corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.Port),
		}

		if port.NodePort > 0 {
			servicePort.NodePort = port.NodePort
		}

		ports = append(ports, servicePort)
	}

	spec := corev1.ServiceSpec{
		Ports: ports,
		Selector: map[string]string{
			"app":   env.GetName(),
			"group": "application",
		},
		Type: corev1.ServiceTypeClusterIP,
	}

	if env.Spec.Network.Service.ExternalName != nil {
		spec.ExternalName = *env.Spec.Network.Service.ExternalName
	}

	return &corev1.Service{
		ObjectMeta: metatdata,
		Spec:       spec,
	}
}

func (b *Builder) buildIngress() *networkingv1.Ingress {
	env := b.observed.Env
	ingress := b.observed.Ingress

	if ingress == nil {
		ingress = &networkingv1.Ingress{}
	}

	metatdata := metav1.ObjectMeta{
		Name:      env.Name,
		Namespace: env.Namespace,
		Annotations: mergeMap(map[string]string{
			"seaway.ctx.sh/revision": env.GetRevision(),
		}, env.Spec.Network.Ingress.Annotations),
		Labels: ingress.Labels,
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	backend := &networkingv1.IngressBackend{
		Service: &networkingv1.IngressServiceBackend{
			Name: env.GetName(),
			Port: networkingv1.ServiceBackendPort{
				Number: *env.Spec.Network.Ingress.Port,
			},
		},
	}

	spec := networkingv1.IngressSpec{
		DefaultBackend: backend,
	}

	if env.Spec.Network.Ingress.ClassName != nil {
		spec.IngressClassName = env.Spec.Network.Ingress.ClassName
	}

	if env.Spec.Network.Ingress.TLS != nil {
		spec.TLS = env.Spec.Network.Ingress.TLS
	}

	return &networkingv1.Ingress{
		ObjectMeta: metatdata,
		Spec:       spec,
	}
}

func mergeMap(source, target map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}

	for k, v := range source {
		target[k] = v
	}

	return target
}

// func mergeContainer(source corev1.Container, target []corev1.Container) []corev1.Container {
// 	if target == nil {
// 		target = make([]corev1.Container, 0)
// 	}

// 	containerMap := make(map[string]corev1.Container)
// 	for _, container := range target {
// 		containerMap[container.Name] = container
// 	}

// 	containerMap[source.Name] = source

// 	containers := make([]corev1.Container, 0)
// 	for _, v := range containerMap {
// 		containers = append(containers, v)
// 	}

// 	return containers
// }

func mergeEnvVar(source, target []corev1.EnvVar) []corev1.EnvVar {
	if target == nil {
		target = make([]corev1.EnvVar, 0)
	}

	listMap := make(map[string]corev1.EnvVar)
	for _, v := range source {
		listMap[v.Name] = v
	}

	for _, v := range target {
		listMap[v.Name] = v
	}

	vars := make([]corev1.EnvVar, 0)
	for _, v := range listMap {
		vars = append(vars, v)
	}

	return vars
}
