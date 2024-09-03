package env

import (
	"fmt"
	"os"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/auth"
	"ctx.sh/seaway/pkg/console"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArchiveKey is the path and name of the archive object.
func ArchiveKey(name string, env *v1beta1.ManifestEnvironmentSpec) string {
	return fmt.Sprintf("%s/%s-%s.tar.gz", *env.Source.S3.Prefix, name, env.Namespace)
}

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

// GetSecret returns a new secret containing the credentials.
func GetSecret(name, namespace string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-user",
			Namespace: namespace,
		},
	}
	secret.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	return secret
}

func GetCredentials() (*auth.Credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		console.Fatal("Unable to determine user home directory")
	}

	var creds *auth.Credentials
	// TODO: Make this configurable.
	filename := home + "/.seaway/creds"

	_, err = os.Stat(filename)
	if err != nil {
		creds, err = auth.NewCredentials(filename)
		if err != nil {
			return nil, err
		}
	} else {
		creds, err = auth.LoadCredentials(filename)
		if err != nil {
			return nil, err
		}
	}

	return creds, nil
}
