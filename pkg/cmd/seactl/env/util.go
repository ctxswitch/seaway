package env

import (
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEnvironment returns a new environment object.
func GetEnvironment(name, namespace string) *v1beta1.Environment {
	env := &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	// TODO: We actually need this with the client at this point because we use
	// the gvk to get the resource interface. Revisit this later and refactor it
	// out.  It's not horrible but it's not great either.
	env.SetGroupVersionKind(v1beta1.SchemeGroupVersion.WithKind("Environment"))

	return env
}
