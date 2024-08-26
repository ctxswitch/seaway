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
	"github.com/spf13/cobra"
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
	// ctx := ctrl.SetupSignalHandler()

	if len(args) != 1 {
		return fmt.Errorf("expected context name")
	}

	var manifest v1beta1.Manifest
	if err := manifest.Load("manifest.yaml"); err != nil {
		console.Fatal("Unable to load manifest")
	}

	_, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build context '%s' not found in the manifest", args[0])
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
