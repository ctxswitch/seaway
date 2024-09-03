package collector

import (
	"context"
	"errors"
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
	Env         *v1beta1.Environment
	Job         *batchv1.Job
	Deployment  *appsv1.Deployment
	Service     *corev1.Service
	Ingress     *networkingv1.Ingress
	UserSecret  *corev1.Secret
	observeTime time.Time
}

func NewObservedState() *ObservedState {
	return &ObservedState{
		Env:         nil,
		Job:         nil,
		Deployment:  nil,
		Service:     nil,
		Ingress:     nil,
		UserSecret:  nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
}

func (o *StateObserver) observe(ctx context.Context, observed *ObservedState) error {
	env, err := o.observeEnvironment(ctx)
	if err != nil || env == nil {
		return err
	}

	observed.Env = env

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
	userSecret, err := o.observeUserSecret(ctx, o.Request.Name+"-user")
	if err != nil {
		return err
	}

	if userSecret == nil {
		return errors.New("user secret required but not found")
	}

	observed.UserSecret = userSecret

	return nil
}

func (o *StateObserver) observeEnvironment(ctx context.Context) (*v1beta1.Environment, error) {
	var env v1beta1.Environment
	if err := o.Client.Get(ctx, o.Request.NamespacedName, &env); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
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

func (o *StateObserver) observeUserSecret(ctx context.Context, name string) (*corev1.Secret, error) {
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