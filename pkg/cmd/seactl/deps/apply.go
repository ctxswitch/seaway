package deps

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ApplyUsage     = "apply [context]"
	ApplyShortDesc = "Apply the dependencies to the target object storage using the configuration context"
	ApplyLongDesc  = `Apply the dependencies to the target object storage based on the configuration context`
)

type Apply struct {
	logLevel int8
}

func NewApply() *Apply {
	return &Apply{}
}

func (a *Apply) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(args) != 1 {
		return fmt.Errorf("expected context name")
	}

	// Load the manifest and grab the appropriate environment
	var manifest v1beta1.Manifest
	if err := manifest.Load("manifest.yaml"); err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build context '%s' not found in the manifest", args[0])
	}

	client, err := kube.NewSeawayClient("", "")
	if err != nil {
		console.Fatal(err.Error())
	}

	// Process the kustomize manifests from the base directory
	krusty, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: "config/overlays/dev",
	})
	if err != nil {
		console.Fatal(err.Error())
	}

	items, err := krusty.Resources()
	if err != nil {
		console.Fatal(err.Error())
	}

	console.Info("Applying dependencies for the '%s' environment", env.Name)

	for _, item := range items {
		expected := item.Resource

		obj := GetObject(expected)
		op, err := client.CreateOrUpdate(ctx, obj, func() error {
			// TODO: Ugly. Fix this.  We need to save the initial state of the
			// object so we can preserve the managed fields after copying the
			// values from the expected object.
			existing := obj.DeepCopyObject().(kube.Object)
			expected.DeepCopyInto(obj)
			kube.PreserveManagedFields(existing, obj)
			return nil
		})
		if err != nil {
			console.Fatal(err.Error())
		}

		api := strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
		var out string
		if obj.GetNamespace() == "" {
			out = fmt.Sprintf("%s/%s", api, obj.GetName())
		} else {
			out = fmt.Sprintf("%s/%s/%s", api, obj.GetNamespace(), obj.GetName())
		}

		switch op {
		case kube.OperationResultNone:
			console.Unchanged(out)
		case kube.OperationResultUpdated:
			console.Updated(out)
		case kube.OperationResultCreated:
			console.Created(out)
		}
	}

	return nil
}

func GetObject(u *unstructured.Unstructured) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetName(u.GetName())
	obj.SetNamespace(u.GetNamespace())
	obj.SetGroupVersionKind(u.GroupVersionKind())

	return obj
}

func (a *Apply) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ApplyUsage,
		Short: ApplyShortDesc,
		Long:  ApplyLongDesc,
		RunE:  a.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&a.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")

	return cmd
}
