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

package main

import (
	"ctx.sh/seaway/pkg/build"
	"ctx.sh/seaway/pkg/cmd/seactl/clean"
	"ctx.sh/seaway/pkg/cmd/seactl/install"
	"ctx.sh/seaway/pkg/cmd/seactl/logs"
	"ctx.sh/seaway/pkg/cmd/seactl/sync"
	"github.com/spf13/cobra"
)

const (
	RootUsage        = "seactl [command] [args...]"
	RootShortDesc    = "CLI utility for managing Seaway development environments."
	RootLongDesc     = `CLI utility for managing Seaway development environments.`
	LogsUsage        = "logs [subcommand]"
	LogsShortDesc    = "Utility to stream logs from the development environment"
	LogsLongDesc     = `Utility to stream logs from the development environment`
	InstallUsage     = "install"
	InstallShortDesc = "Installs the controller, CRDs, and optional dependencies."
	InstallLongDesc  = `Installs the controller, CRDs, and optional dependencies.`
	SyncUsage        = "sync"
	SyncShortDesc    = "Sync to the target object storage using the configuration context"
	SyncLongDesc     = `Sync the code to the target object storage based on the configuration context
provided in the manifest.  This will trigger a new development deployment if there was a change.`
	CleanUsage     = "clean"
	CleanShortDesc = "Clean all development environment resources."
	CleanLongDesc  = `Cleans all development environment resources for the specified context.`

	DefaultInstallCrds        = true
	DefaultInstallCertManager = false
	DefaultInstallLocalstack  = false
	DefaultInstallRegistry    = false
	DefaultEnableDevMode      = false
	DefaultSyncOnlyDeps       = false
	DefaultSyncWithDeps       = false
	DefaultSyncForce          = false
)

type Root struct{}

func NewRoot() *Root {
	return &Root{}
}

func (r *Root) Execute() error {
	return r.Command().Execute()
}

func (r *Root) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     RootUsage,
		Short:   RootShortDesc,
		Long:    RootLongDesc,
		Version: build.Version,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	rootCmd.AddCommand(SyncCommand())
	rootCmd.AddCommand(CleanCommand())
	rootCmd.AddCommand(LogsCommand())
	rootCmd.AddCommand(InstallCommand())

	rootCmd.PersistentFlags().StringP("context", "", "", "set the Kubernetes context")
	return rootCmd
}

func LogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   LogsUsage,
		Short: LogsShortDesc,
		Long:  LogsLongDesc,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.AddCommand(logs.NewAppLogs().Command())
	cmd.AddCommand(logs.NewBuildLogs().Command())
	return cmd
}

func InstallCommand() *cobra.Command {
	installer := install.Command{}

	cmd := &cobra.Command{
		Use:   InstallUsage,
		Short: InstallShortDesc,
		Long:  InstallLongDesc,
		RunE:  installer.RunE,
	}

	cmd.PersistentFlags().BoolVarP(&installer.InstallCrds, "install-crds", "", DefaultInstallCrds, "enable crd installation")
	cmd.PersistentFlags().BoolVarP(&installer.InstallCertManager, "install-cert-manager", "", DefaultInstallCertManager, "enable cert manager installation")
	cmd.PersistentFlags().BoolVarP(&installer.InstallLocalStack, "install-localstack", "", DefaultInstallLocalstack, "enable localstack installation for object storage")
	cmd.PersistentFlags().BoolVarP(&installer.EnableDevMode, "enable-dev-mode", "", DefaultEnableDevMode, "enable seaway development mode")
	cmd.PersistentFlags().BoolVarP(&installer.InstallRegistry, "install-registry", "", DefaultInstallRegistry, "enable seaway container registry installation")
	return cmd
}

func SyncCommand() *cobra.Command {
	s := sync.Command{}
	cmd := &cobra.Command{
		Use:   SyncUsage,
		Short: SyncShortDesc,
		Long:  SyncLongDesc,
		RunE:  s.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&s.LogLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().BoolVarP(&s.Force, "force", "", DefaultSyncForce, "force a resync even if no changes are detected")
	cmd.PersistentFlags().BoolVarP(&s.WithDeps, "with-deps", "", DefaultSyncWithDeps, "apply dependencies before syncing the application")
	cmd.PersistentFlags().BoolVarP(&s.OnlyDeps, "only-deps", "", DefaultSyncOnlyDeps, "apply the dependencies without syncing the application")
	return cmd
}

func CleanCommand() *cobra.Command {
	c := clean.Command{}

	cmd := &cobra.Command{
		Use:   CleanUsage,
		Short: CleanShortDesc,
		Long:  CleanLongDesc,
		RunE:  c.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&c.LogLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")

	return cmd
}
