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

	"ctx.sh/seaway/pkg/auth"
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
	context         string
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
	if _, err := exec.LookPath("kubectl"); err != nil {
		console.Fatal("kubectl is not installed")
	}

	if _, err := exec.LookPath("kustomize"); err != nil {
		console.Fatal("kustomize is not installed")
	}

	if _, err := exec.LookPath("envsubst"); err != nil {
		console.Fatal("envsubst is not installed")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		console.Fatal("Unable to determine user home directory")
	}

	var creds *auth.Credentials
	filename := home + "/.seaway/creds"
	if _, err := os.Stat(filename); err != nil {
		creds, err = auth.NewCredentials(filename)
		if err != nil {
			console.Fatal("Unable to create credentials file: %v", err)
		}
	} else {
		creds, err = auth.LoadCredentials(filename)
		if err != nil {
			console.Fatal("Unable to load credentials file: %v", err)
		}
	}

	if err := os.Setenv("SEAWAY_S3_ACCESS_KEY", creds.GetAccessKey()); err != nil {
		console.Fatal("Unable to set SEAWAY_S3_ACCESS_KEY")
	}

	if err := os.Setenv("SEAWAY_S3_SECRET_KEY", creds.GetSecretKey()); err != nil {
		console.Fatal("Unable to set SEAWAY_S3_SECRET_KEY")
	}

	if err := os.Setenv("MINIO_ROOT_USER", creds.GetMinioRootUser()); err != nil {
		console.Fatal("Unable to set MINIO_ROOT_USER")
	}

	if err := os.Setenv("MINIO_ROOT_PASSWORD", creds.MinioRootPassword); err != nil {
		console.Fatal("Unable to set MINIO_ROOT_PASSWORD")
	}

	if err := os.Setenv("CONTEXT", "--context="+s.context); err != nil {
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
	cmd.PersistentFlags().StringVarP(&s.context, "context", "c", "", "specify the kubernetes context to use")
	cmd.PersistentFlags().BoolVarP(&s.enableTailscale, "enable-tailscale", "", false, "enable tailscale ingress controller on the cluster")
	return cmd
}
