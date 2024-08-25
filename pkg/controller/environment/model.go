package environment

import (
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetEnvironmentJob(env *v1beta1.Environment, scheme *runtime.Scheme, etag string) batchv1.Job {
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      env.GetName() + "-" + etag,
			Namespace: env.GetNamespace(),
		},
	}

	controllerutil.SetControllerReference(env, &job, scheme) //nolint:errcheck

	return job
}

func GetEnvironmentDeployment(env *v1beta1.Environment, scheme *runtime.Scheme) appsv1.Deployment {
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      env.GetName(),
			Namespace: env.GetNamespace(),
		},
	}

	controllerutil.SetControllerReference(env, &deploy, scheme) //nolint:errcheck

	return deploy
}
