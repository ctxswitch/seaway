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
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
	"github.com/spf13/cobra"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	watchtools "k8s.io/client-go/tools/watch"
)

const (
	ApplyUsage     = "apply [context]"
	ApplyShortDesc = "Apply the dependencies to the target object storage using the configuration context"
	ApplyLongDesc  = `Apply the dependencies to the target object storage based on the configuration context`
)

var (
	DefaultBackoff = wait.Backoff{ //nolint:gochecknoglobals
		Steps:    5,
		Duration: 200 * time.Millisecond,
		Factor:   2.0,
	}
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
		a.do(ctx, client, dep)
	}

	return nil
}

func (a *Apply) do(ctx context.Context, client *kube.KubectlCmd, dep v1beta1.ManifestDependency) error {
	// Process the kustomize manifests from the base directory
	krusty, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: dep.Path,
	})
	if err != nil {
		console.Fatal("Unable to initialize kustomize: %s", err.Error())
		return err
	}

	err = krusty.Build()
	if err != nil {
		console.Fatal("Unable to perform kustomize build: %s", err.Error())
		return err
	}

	items := krusty.Resources()

	console.Info("Applying dependency '%s'", dep.Path)
	for _, item := range items {
		if err := a.apply(ctx, client, item); err != nil {
			console.Fatal("error: %s", err.Error())
		}
	}

	if len(dep.Wait) == 0 {
		return nil
	}

	console.Info("Waiting for resource conditions")
	for _, cond := range dep.Wait {
		if obj, ok := krusty.GetResource(cond.Kind, cond.Name); ok {
			err := a.wait(ctx, client, obj, cond)
			if err != nil {
				console.Fatal("error: %s", err.Error())
			}
		}
	}

	return nil
}

func (a *Apply) apply(ctx context.Context, client *kube.KubectlCmd, k kustomize.KustomizerResource) error {
	obj := k.Resource.DeepCopy()
	api := ToAPIString(obj)

	var op kube.OperationResult
	var err error

	err = wait.ExponentialBackoffWithContext(ctx, DefaultBackoff, func(context.Context) (bool, error) {
		op, err = client.CreateOrUpdate(ctx, obj, func() error {
			return nil
		})
		if err == nil {
			return true, nil
		}

		if apierr.IsNotFound(err) {
			return false, nil
		}

		return false, err
	})
	if err != nil {
		console.Fatal("error applying resource %s: %s", api, err.Error())
		return err
	}

	switch op {
	case kube.OperationResultNone:
		console.Unchanged(api)
	case kube.OperationResultUpdated:
		console.Updated(api)
	case kube.OperationResultCreated:
		console.Created(api)
	}

	return nil
}

type KindName struct {
	Kind string
	Name string
}

func (a *Apply) wait(ctx context.Context, client *kube.KubectlCmd, k kustomize.KustomizerResource, cond v1beta1.ManifestWaitCondition) error {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, cond.Timeout)
	defer cancel()

	obj := k.Resource

	out := ToAPIString(obj)
	console.Waiting(out)

	err := client.WaitForCondition(ctx, obj, cond.For, cond.Timeout)
	if err != nil {
		return err
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

// ToAPIString returns the gvk, namespace, and name of the object as a string.
func ToAPIString(obj *unstructured.Unstructured) string {
	api := strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
	var out string
	if obj.GetNamespace() == "" {
		out = fmt.Sprintf("%s/%s", api, obj.GetName())
	} else {
		out = fmt.Sprintf("%s/%s/%s", api, obj.GetNamespace(), obj.GetName())
	}

	return out
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
