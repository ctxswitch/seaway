package install

import (
	"context"
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/cmd/util"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os/signal"
	"syscall"
)

type Command struct {
	InstallCrds        bool
	InstallCertManager bool
	InstallLocalStack  bool
	InstallRegistry    bool
	EnableDevMode      bool
}

// nolint:funlen,gocognit
func (c *Command) RunE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	console.Section("Installing Seaway")

	if c.InstallCrds {
		console.Info("Installing CRDs")
		raw, err := GetCrdBytes()
		if err != nil {
			return err
		}

		err = c.install(ctx, raw, []v1beta1.ManifestWaitCondition{})
		if err != nil {
			return err
		}
	}

	if c.InstallCertManager {
		console.Info("Installing cert-manager")
		raw, err := GetCertManagerBytes()
		if err != nil {
			return err
		}

		err = c.install(ctx, raw, []v1beta1.ManifestWaitCondition{
			{
				Kind: "Deployment",
				Name: "cert-manager",
				For:  "ready",
			},
			{
				Kind: "Deployment",
				Name: "cert-manager-webhook",
				For:  "ready",
			},
		})
		if err != nil {
			return err
		}
	}

	if c.InstallLocalStack {
		console.Info("Installing localstack")
		raw, err := GetLocalstackBytes()
		if err != nil {
			return err
		}

		err = c.install(ctx, raw, []v1beta1.ManifestWaitCondition{
			{
				Kind: "Deployment",
				Name: "localstack",
				For:  "ready",
			},
		})
		if err != nil {
			return err
		}
	}

	if c.InstallRegistry {
		console.Info("Installing registry")
		raw, err := GetRegistryBytes()
		if err != nil {
			return err
		}

		err = c.install(ctx, raw, []v1beta1.ManifestWaitCondition{
			{
				Kind: "Deployment",
				Name: "registry",
				For:  "ready",
			},
		})
		if err != nil {
			return err
		}
	}

	console.Info("Installing the seaway controller")
	raw, err := GetControllerBytes()
	if err != nil {
		return err
	}

	// TODO: do something better here.  It's wasteful to unpack and repack
	raw, err = c.modifyController(raw)
	if err != nil {
		return err
	}

	err = c.install(ctx, raw, []v1beta1.ManifestWaitCondition{
		{
			Kind: "Deployment",
			Name: "seaway-controller",
			For:  "ready",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Command) install(ctx context.Context, raw []byte, waitConditions []v1beta1.ManifestWaitCondition) error {
	// Take the bytes and build them into the kustomize resources
	// Use the approach that we take in deps to install then wait, if we are waiting
	krusty, err := kustomize.NewKustomizerFromBytes(raw)
	if err != nil {
		return err
	}

	err = krusty.Build()
	if err != nil {
		return err
	}

	client, err := kube.NewKubectlCmd("", "")
	if err != nil {
		return err
	}

	for _, item := range krusty.Resources() {
		if err := util.Apply(ctx, client, item); err != nil {
			console.Fatal("error: %s", err.Error())
		}
	}

	if len(waitConditions) == 0 {
		return nil
	}

	console.Info("Waiting for resource conditions")
	for _, cond := range waitConditions {
		if obj, ok := krusty.GetResource(cond.Kind, cond.Name); ok {
			err := util.Wait(ctx, client, obj, cond)
			if err != nil {
				console.Fatal("error: %s", err.Error())
			}
		}
	}

	return nil
}

func (c *Command) modifyController(raw []byte) ([]byte, error) {
	krusty, err := kustomize.NewKustomizerFromBytes(raw)
	if err != nil {
		return nil, err
	}

	err = krusty.Build()
	if err != nil {
		return nil, err
	}

	resource, ok := krusty.GetResource("Deployment", "seaway-controller")
	if !ok {
		return nil, fmt.Errorf("seaway controller deployment not found")
	}

	obj := resource.Resource.Object
	containersIface, _, err := unstructured.NestedFieldCopy(obj, "spec", "template", "spec", "containers")
	if err != nil {
		return nil, err
	}

	volumesIface, _, err := unstructured.NestedFieldCopy(obj, "spec", "template", "spec", "volumes")
	if err != nil {
		return nil, err
	}

	containers, _ := containersIface.([]interface{})
	container, _ := containers[0].(map[string]interface{})
	volumes, _ := volumesIface.([]interface{})

	if c.EnableDevMode { // nolint:nestif
		// TODO: This is assuming that we will always have a single container.  It's a safe
		//       assumption right now, but may not be in the future.  We should make this a
		//       little more resilient.
		err = unstructured.SetNestedField(container, "docker.io/golang:latest", "image")
		if err != nil {
			return nil, err
		}

		err = unstructured.SetNestedStringSlice(container, []string{"sleep", "infinity"}, "command")
		if err != nil {
			return nil, err
		}

		err = unstructured.SetNestedField(container, "/usr/src/app", "workingDir")
		if err != nil {
			return nil, err
		}

		volumeMountsIface, _, verr := unstructured.NestedFieldNoCopy(container, "volumeMounts")
		if verr != nil {
			return nil, verr
		}

		volumeMounts, _ := volumeMountsIface.([]interface{})
		volumeMounts = append(volumeMounts, map[string]interface{}{
			"name":      "app",
			"mountPath": "/usr/src/app",
			"readOnly":  true,
		})

		err = unstructured.SetNestedField(container, volumeMounts, "volumeMounts")
		if err != nil {
			return nil, err
		}

		volumes = append(volumes, map[string]interface{}{
			"name": "app",
			"hostPath": map[string]interface{}{
				"path": "/app",
			},
		})
	}

	err = unstructured.SetNestedField(obj, containers, "spec", "template", "spec", "containers")
	if err != nil {
		return nil, err
	}

	err = unstructured.SetNestedField(obj, volumes, "spec", "template", "spec", "volumes")
	if err != nil {
		return nil, err
	}

	err = krusty.SetResource(&unstructured.Unstructured{Object: obj})
	if err != nil {
		return nil, err
	}

	return krusty.ToYamlBytes()
}
