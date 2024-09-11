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

package init

import (
	"os"
	"os/exec"
	"strconv"

	"ctx.sh/seaway/pkg/build"
	"ctx.sh/seaway/pkg/console"
	"github.com/spf13/cobra"
)

const (
	SharedUsage     = "shared"
	SharedShortDesc = "Initializes seaway resources for a kubernetes cluster."
	SharedLongDesc  = `Initializes seaway resources for a kubernetes cluster.`
)

type Shared struct {
	logLevel        int8
	enableTailscale bool
}

func NewShared() *Shared {
	return &Shared{}
}

// RunE is the main function for the init command which installs all required
// seaway resources on a cluster.  This is a ***very*** temporary solution to
// keep the install as simple as possible for users.  In the future, there will
// be no external dependencies and the install will be done completely by seactl.
func (s *Shared) RunE(cmd *cobra.Command, args []string) error {
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()

	if _, err := exec.LookPath("kubectl"); err != nil {
		console.Fatal("kubectl is not installed")
	}

	if _, err := exec.LookPath("kustomize"); err != nil {
		console.Fatal("kustomize is not installed")
	}

	if _, err := exec.LookPath("envsubst"); err != nil {
		console.Fatal("envsubst is not installed")
	}

	// TODO: Just create the secrets later.  Right now we rely on them being set
	// when we replace them in the script.
	if os.Getenv("SEAWAY_S3_ACCESS_KEY") == "" {
		console.Fatal("SEAWAY_S3_ACCESS_KEY is not set")
	}

	if os.Getenv("SEAWAY_S3_SECRET_KEY") == "" {
		console.Fatal("SEAWAY_S3_SECRET_KEY is not set")
	}

	if os.Getenv("MINIO_ROOT_USER") == "" {
		console.Fatal("MINIO_ROOT_USER is not set")
	}

	if os.Getenv("MINIO_ROOT_PASSWORD") == "" {
		console.Fatal("MINIO_ROOT_PASSWORD is not set")
	}

	if err := os.Setenv("CONTEXT", "--context="+kubeContext); err != nil {
		console.Fatal("Unable to set CONTEXT")
	}

	if err := os.Setenv("SEAWAY_VERSION", build.Version); err != nil {
		console.Fatal("Unable to set SEAWAY_VERSION")
	}

	enabled := strconv.FormatBool(s.enableTailscale)
	if err := os.Setenv("ENABLE_TAILSCALE", enabled); err != nil {
		console.Fatal("Unable to set ENABLE_TAILSCALE")
	}

	script := exec.Command("bash", "-c", sharedScript)
	script.Stderr = os.Stderr
	script.Stdout = os.Stdout
	if err := script.Start(); err != nil {
		console.Fatal("Error starting script: %v", err)
	}

	if err := script.Wait(); err != nil {
		console.Fatal("Error running script: %v", err)
	}

	return nil
}

// Command creates the clean command.
func (s *Shared) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   SharedUsage,
		Short: SharedShortDesc,
		Long:  SharedLongDesc,
		RunE:  s.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&s.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().BoolVarP(&s.enableTailscale, "enable-tailscale", "", false, "enable tailscale ingress controller on the cluster")
	return cmd
}
