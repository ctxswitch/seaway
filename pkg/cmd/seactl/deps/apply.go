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
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/cmd/util"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
	"fmt"
	"github.com/spf13/cobra"
	"os/signal"
	"syscall"
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

	console.Section("Applying dependencies for the '%s' environment", env.Name)

	for _, dep := range env.Dependencies {
		// TODO: pull the errors back to here.
		_ = a.do(ctx, client, dep)
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
		if err := util.Apply(ctx, client, item); err != nil {
			console.Fatal("error: %s", err.Error())
		}
	}

	if len(dep.Wait) == 0 {
		return nil
	}

	console.Info("Waiting for resource conditions")
	for _, cond := range dep.Wait {
		if obj, ok := krusty.GetResource(cond.Kind, cond.Name); ok {
			err := util.Wait(ctx, client, obj, cond)
			if err != nil {
				console.Fatal("error: %s", err.Error())
			}
		}
	}

	return nil
}

type KindName struct {
	Kind string
	Name string
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
