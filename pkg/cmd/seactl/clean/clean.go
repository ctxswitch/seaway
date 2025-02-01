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

package clean

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/cmd/util"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Command struct {
	LogLevel int8
}

// RunE is the main function for the clean command which removes all artifacts
// and objects associated with a development environment.
// TODO: Implement the clean command.
func (c *Command) RunE(cmd *cobra.Command, args []string) error {
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(args) != 1 {
		return fmt.Errorf("expected environment name")
	}

	var manifest v1beta1.Manifest
	err := manifest.Load("manifest.yaml")
	if err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build environment '%s' not found in the manifest", args[0])
	}

	console.Info("Cleaning environment '%s'", env.Name)

	client, err := kube.NewKubectlCmd("", kubeContext)
	if err != nil {
		console.Fatal(err.Error())
	}

	obj := util.GetEnvironment(manifest.Name, env.Namespace)
	err = client.Delete(ctx, obj, metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			console.Fatal("Unable to delete environment: %s", err)
		}
	}

	console.Info("Deleting source archive")

	// TODO: Need a delete endpoint for the controller or force the controller to delete the
	// archive when the environment is deleted as part of the finalizer.

	return delete(ctx, client, env)
}
