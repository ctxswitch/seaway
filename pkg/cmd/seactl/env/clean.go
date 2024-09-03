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

package env

import (
	"fmt"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/storage"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	CleanUsage     = "clean [context]"
	CleanShortDesc = "Clean all development environment resources."
	CleanLongDesc  = `Cleans all development environment resources for the specified context.`
)

type Clean struct {
	logLevel int8
}

func NewClean() *Clean {
	return &Clean{}
}

// RunE is the main function for the clean command which removes all artifacts
// and objects associated with a development environment.
// TODO: Implement the clean command.
func (c Clean) RunE(cmd *cobra.Command, args []string) error {
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()

	ctx := ctrl.SetupSignalHandler()

	if len(args) != 1 {
		return fmt.Errorf("expected environment name")
	}

	creds, err := GetCredentials()
	if err != nil {
		console.Fatal("Unable to get credentials: %s", err)
	}

	var manifest v1beta1.Manifest
	err = manifest.Load("manifest.yaml")
	if err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build environment '%s' not found in the manifest", args[0])
	}

	console.Info("Cleaning environment '%s'", env.Name)

	client, err := kube.NewSeawayClient("", kubeContext)
	if err != nil {
		console.Fatal(err.Error())
	}

	obj := GetEnvironment(manifest.Name, env.Namespace)
	err = client.Delete(ctx, obj, metav1.DeleteOptions{})
	if err != nil {
		console.Fatal("Unable to delete environment: %s", err)
	}

	console.Info("Deleting source archive")

	store := storage.NewClient(env.Store.GetEndpoint(), env.Store.UseSSL())
	err = store.Connect(ctx, creds)
	if err != nil {
		console.Fatal("Unable to connect to object storage: %s", err)
	}

	bucket := *env.Store.Bucket
	key := ArchiveKey(manifest.Name, &env)
	err = store.DeleteObject(ctx, bucket, key)
	if err != nil {
		console.Fatal("Unable to delete source archive: %s", err)
	}

	return nil
}

// Command creates the clean command.
func (c *Clean) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   CleanUsage,
		Short: CleanShortDesc,
		Long:  CleanLongDesc,
		RunE:  c.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&c.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")

	return cmd
}
