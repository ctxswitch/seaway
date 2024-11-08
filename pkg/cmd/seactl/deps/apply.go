// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// RunE generates and applies the dependencies for a development environment.
func (a *Apply) RunE(cmd *cobra.Command, args []string) error {
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()
	client, err := kube.NewKubectlCmd("", kubeContext)
	if err != nil {
		console.Fatal(err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(args) != 1 {
		return fmt.Errorf("expected context name")
	}

	// Load the manifest and grab the appropriate environment
	var manifest v1beta1.Manifest
	err = manifest.Load("manifest.yaml")
	if err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build environment '%s' not found in the manifest", args[0])
	}

	console.Info("Applying dependencies for the '%s' environment", env.Name)

	for _, dep := range env.Dependencies {
		console.Info("Applying dependency '%s'", dep.Path)
		if err := a.apply(ctx, client, dep.Path); err != nil {
			console.Fatal("unable to apply: %s", err.Error())
		}
	}

	return nil
}

func (a *Apply) apply(ctx context.Context, client *kube.KubectlCmd, path string) error {
	// Process the kustomize manifests from the base directory
	krusty, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path,
	})
	if err != nil {
		console.Fatal("Unable to initialize kustomize: %s", err.Error())
		return err
	}

	items, err := krusty.Resources()
	if err != nil {
		console.Fatal("resources: ", err.Error())
		return err
	}

	for _, item := range items {
		expected := item.Resource

		// TODO: make a dry-run option
		obj := GetObject(expected)
		op, err := client.CreateOrUpdate(ctx, obj, func() error {
			// TODO: Ugly. Fix this.  We need to save the initial state of the
			// object so we can preserve the api managed fields after copying the
			// values from the expected object.
			existing, can := obj.DeepCopyObject().(kube.Object)
			if !can {
				return fmt.Errorf("could not cast existing object")
			}
			expected.DeepCopyInto(obj)
			kube.PreserveManagedFields(existing, obj)

			return nil
		})
		if err != nil {
			api := strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
			console.Fatal("error applying resource %s/%s: %s", api, obj.GetName(), err.Error())
			return err
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

// GetObject returns a new unstructured object from the provided object.  Because our client
// utilizes the dynamic client interface we need to ensure that the GVK is also set so we
// can properly set up the resource interface.
func GetObject(u *unstructured.Unstructured) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetName(u.GetName())
	obj.SetNamespace(u.GetNamespace())
	obj.SetGroupVersionKind(u.GroupVersionKind())

	return obj
}

// Command returns the cobra command for the apply subcommand.
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
