package collector

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObservedState struct {
	Env                *v1beta1.Environment
	Job                *batchv1.Job
	Deployment         *appsv1.Deployment
	Service            *corev1.Service
	Ingress            *networkingv1.Ingress
	StorageCredentials *corev1.Secret
	EnvCredentials     *corev1.Secret
	Config             *v1beta1.EnvironmentConfig
	observeTime        time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		Env:         nil,
		Job:         nil,
		Deployment:  nil,
		Service:     nil,
		Ingress:     nil,
		Config:      nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	env, err := o.observeEnvironment(ctx)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil
		}
		return err
	}

	observed.Env = env

	// TODO: Allow user to define the namespace.
	config, err := o.observeConfig(ctx, env, v1beta1.DefaultControllerNamespace)
	if err != nil {
		// Handle missing error more gracefully
		return err
	}

	observed.Config = config

	storageCredentials, err := o.observeStorageCredentials(ctx, config.Spec.EnvironmentConfigStorageSpec.Credentials, config.GetNamespace())
	if err != nil {
		return err
	}

	observed.StorageCredentials = storageCredentials

	// Observe the job
	job, err := o.observeJob(ctx, o.Request.Name+"-build")
	if err != nil {
		return err
	}

	observed.Job = job

	// Observe the deployment
	deployment, err := o.observeDeployment(ctx)
	if err != nil {
		return err
	}

	observed.Deployment = deployment

	// Observe the service
	service, err := o.observeService(ctx)
	if err != nil {
		return err
	}

	observed.Service = service

	// Observe the ingress
	ingress, err := o.observeIngress(ctx)
	if err != nil {
		return err
	}

	observed.Ingress = ingress

	// Observe the user secret
	// TODO: This is going to end up being a copy of the storage credentials that
	// lives in the environment namespace.  I'm not a huge fan of this from a security/
	// isolation perspective, but it's the easiest way to keep the build jobs local to
	// the environment namespace at the moment.  The alternative is just inject the
	// creds directly into the build job's environment - which is even worse though
	// I'm probably overthinking it at this early stage.
	credentials, err := o.observeEnvCredentials(ctx, env.Name+"-credentials")
	if err != nil {
		return err
	}

	observed.EnvCredentials = credentials

	return nil
}

func (o *StateObserver) observeEnvironment(ctx context.Context) (*v1beta1.Environment, error) {
	var env v1beta1.Environment
	if err := o.Client.Get(ctx, o.Request.NamespacedName, &env); err != nil {
		return nil, err
	}
	v1beta1.Defaulted(&env)
	return &env, nil
}

func (o *StateObserver) observeJob(ctx context.Context, name string) (*batchv1.Job, error) {
	var job batchv1.Job
	if err := o.Client.Get(ctx, types.NamespacedName{
		Namespace: o.Request.Namespace,
		Name:      name,
	}, &job); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (o *StateObserver) observeDeployment(ctx context.Context) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := o.Client.Get(ctx, o.Request.NamespacedName, &deployment); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return nil, err
	}
	return &deployment, nil
}

func (o *StateObserver) observeService(ctx context.Context) (*corev1.Service, error) {
	var service corev1.Service
	if err := o.Client.Get(ctx, o.Request.NamespacedName, &service); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return nil, err
	}
	return &service, nil
}

func (o *StateObserver) observeIngress(ctx context.Context) (*networkingv1.Ingress, error) {
	var ingress networkingv1.Ingress
	if err := o.Client.Get(ctx, o.Request.NamespacedName, &ingress); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return nil, err
	}
	return &ingress, nil
}

func (o *StateObserver) observeEnvCredentials(ctx context.Context, name string) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := o.Client.Get(ctx, types.NamespacedName{
		Namespace: o.Request.Namespace,
		Name:      name,
	}, &secret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

func (o *StateObserver) observeStorageCredentials(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	if name == "" {
		// If there are no storage credentials defined, create our own anonymous secret
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "anonymous",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"AWS_ACCESS_KEY_ID":     []byte("anonymous"),
				"AWS_SECRET_ACCESS_KEY": []byte("anonymous"),
			},
		}, nil
	}

	var secret corev1.Secret
	if err := o.Client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &secret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, fmt.Errorf("storage credentials not found: %s/%s", name, namespace)
		}
		return nil, err
	}
	return &secret, nil
}

func (o *StateObserver) observeConfig(ctx context.Context, env *v1beta1.Environment, namespace string) (*v1beta1.EnvironmentConfig, error) {
	// TODO: How does this impact a multiuser environment?
	name := env.GetName()
	if env.Spec.Config != "" {
		name = env.Spec.Config
	}

	var config v1beta1.EnvironmentConfig
	if err := o.Client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, &config); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, fmt.Errorf("seaway config not found: %s/%s", name, namespace)
		}
		return nil, err
	}
	return &config, nil
}
