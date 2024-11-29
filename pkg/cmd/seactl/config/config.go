package config

import (
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Command struct {
	Namespace             string
	StorageEndpoint       string
	StorageBucket         string
	StorageRegion         string
	StorageForcePathStyle bool
	StorageCredentials    string
	StoragePrefix         string
	RegistryURL           string
	RegistryNodeport      int32
}

func (c *Command) RunE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		console.Fatal("Exactly one argument expected: <name>")
	}

	name := args[0]
	console.Info("Configuration for %s", name)

	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()

	client, err := kube.NewKubectlCmd("", kubeContext)
	if err != nil {
		console.Fatal("unable to create kubernetes client: %s", err.Error())
		return err
	}

	obj := GetSeawayConfig(name, c.Namespace)

	op, err := client.CreateOrUpdate(cmd.Context(), obj, func() error {
		// TODO: With the way this command is structured, it would be nice to only
		//       update the values that are passed at the time and only set the defaults
		//       if the object is being created.
		storage := v1beta1.EnvironmentConfigStorageSpec{
			Bucket:         c.StorageBucket,
			Endpoint:       c.StorageEndpoint,
			Prefix:         c.StoragePrefix,
			Region:         c.StorageRegion,
			Credentials:    c.StorageCredentials,
			ForcePathStyle: c.StorageForcePathStyle,
		}

		registry := v1beta1.EnvironmentConfigRegistrySpec{
			URL:      c.RegistryURL,
			NodePort: c.RegistryNodeport,
		}

		obj.Spec.EnvironmentConfigStorageSpec = storage
		obj.Spec.EnvironmentConfigRegistrySpec = registry

		return nil
	})
	if err != nil {
		console.Fatal("Unable to create the seaway configuration", err)
	}

	switch op {
	case kube.OperationResultNone:
		console.Unchanged("Unchanged")
	case kube.OperationResultUpdated:
		console.Updated("Updated")
	case kube.OperationResultCreated:
		console.Created("Created")
	}

	console.Info("Storage Details")
	console.ListNotice("Endpoint: %s", obj.Spec.EnvironmentConfigStorageSpec.Endpoint)
	console.ListNotice("Bucket: %s", obj.Spec.EnvironmentConfigStorageSpec.Bucket)
	console.ListNotice("Region: %s", obj.Spec.EnvironmentConfigStorageSpec.Region)
	console.ListNotice("Prefix: %s", obj.Spec.EnvironmentConfigStorageSpec.Prefix)
	console.ListNotice("Credentials: %s", obj.Spec.EnvironmentConfigStorageSpec.Credentials)
	console.ListNotice("ForcePathStyle: %t", obj.Spec.EnvironmentConfigStorageSpec.ForcePathStyle)

	console.Info("Registry Details")
	console.ListNotice("Registry URL: %s", obj.Spec.EnvironmentConfigRegistrySpec.URL)
	console.ListNotice("Registry Nodeport: %d", obj.Spec.EnvironmentConfigRegistrySpec.NodePort)

	return nil
}

// GetSeawayConfig returns a new seaway configuration object.
func GetSeawayConfig(name, namespace string) *v1beta1.EnvironmentConfig {
	env := &v1beta1.EnvironmentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	env.SetGroupVersionKind(v1beta1.SchemeGroupVersion.WithKind("EnvironmentConfig"))
	return env
}
