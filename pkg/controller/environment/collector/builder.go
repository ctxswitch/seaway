package collector

import (
	"fmt"
	"net/url"
	"strconv"

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
	Job        *batchv1.Job
	Deployment *appsv1.Deployment
	Service    *corev1.Service
	Ingress    *networkingv1.Ingress
}

func NewDesiredState() *DesiredState {
	return &DesiredState{
		Job:        nil,
		Deployment: nil,
		Service:    nil,
		Ingress:    nil,
	}
}

type Builder struct {
	observed *ObservedState
	scheme   *runtime.Scheme
	nodePort int32
	registry *url.URL
}

func (b *Builder) desired(d *DesiredState) error {
	env := b.observed.Env
	if env == nil {
		return nil
	}

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

func (b *Builder) buildJob() *batchv1.Job { //nolint:funlen
	env := b.observed.Env
	job := b.observed.Job

	if job == nil {
		job = &batchv1.Job{}
	}

	metatdata := metav1.ObjectMeta{
		Name:      env.Name + "-build",
		Namespace: env.Namespace,
		Annotations: mergeMap(map[string]string{
			"seaway.ctx.sh/revision": env.GetRevision(),
		}, job.Annotations),
		OwnerReferences: []metav1.OwnerReference{
			env.GetControllerReference(),
		},
	}

	args := env.Spec.Build.Args
	if args == nil {
		args = []string{
			fmt.Sprintf("--dockerfile=%s", *env.Spec.Build.Dockerfile),
			fmt.Sprintf("--context=s3://%s/%s", env.GetBucket(), env.GetKey()),
			fmt.Sprintf("--destination=%s/%s:%s", b.registry.Host, env.GetName(), env.GetRevision()),
			// TODO: toggle caching
			"--cache=true",
			fmt.Sprintf("--cache-repo=%s/build-cache", b.registry.Host),
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
			Value: *env.Spec.Store.Region,
		},
		{
			Name: "S3_ENDPOINT",
			// Need to add the protocol...  Either force it and strip it when setting
			// up the client or add it here.
			Value: "http://" + *env.Spec.Store.Endpoint,
		},
		{
			Name:  "S3_FORCE_PATH_STYLE",
			Value: strconv.FormatBool(*env.Spec.Store.ForcePathStyle),
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
						Name: b.observed.UserSecret.GetName(),
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
		Image:          fmt.Sprintf("localhost:%d/%s:%s", b.nodePort, env.GetName(), env.GetRevision()),
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
				// TODO: Right now just take the first port. I need to allow building out for
				// multiple ports in the future, but for a quick and dirty implementation,
				// this will work for now.  Just doc it and move on.
				Number: *env.Spec.Network.Ingress.Port,
			},
		},
	}
	// TODO: Messy.  Don't like the coupling here.
	// if env.Spec.Networking.Ingress.Enabled && env.Spec.Networking.Service.Enabled {
	// 	backend.Service = &networkingv1.IngressServiceBackend{
	// 		Name: env.GetName(),
	// 		Port: networkingv1.ServiceBackendPort{
	// 			// TODO: Right now just take the first port. I need to allow building out for
	// 			// multiple ports in the future, but for a quick and dirty implementation,
	// 			// this will work for now.  Just doc it and move on.
	// 			Number: *env.Spec.Networking.Ingress.Port,
	// 		},
	// 	}
	// }

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
