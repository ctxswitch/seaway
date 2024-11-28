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
	depscmd "ctx.sh/seaway/pkg/cmd/seactl/deps"
	envcmd "ctx.sh/seaway/pkg/cmd/seactl/env"
	"ctx.sh/seaway/pkg/cmd/seactl/install"
	"ctx.sh/seaway/pkg/cmd/seactl/logs"
	"github.com/spf13/cobra"
)

const (
	RootUsage        = "seactl [command] [args...]"
	RootShortDesc    = "CLI utility for managing Seaway development environments."
	RootLongDesc     = `CLI utility for managing Seaway development environments.`
	DepsUsage        = "deps [subcommand] [context]"
	DepsShortDesc    = "Utility to manage application dependencies"
	DepsLongDesc     = `Utility to manage application dependencies`
	EnvUsage         = "env [subcommand] [context]"
	EnvShortDesc     = "Utility to manage development environments"
	EnvLongDesc      = `Utility to manage development environments`
	LogsUsage        = "logs [subcommand]"
	LogsShortDesc    = "Utility to stream logs from the development environment"
	LogsLongDesc     = `Utility to stream logs from the development environment`
	InstallUsage     = "install"
	InstallShortDesc = "Installs the controller, CRDs, and optional dependencies."
	InstallLongDesc  = `Installs the controller, CRDs, and optional dependencies.`

	DefaultInstallCrds        = true
	DefaultInstallCertManager = false
	DefaultInstallLocalstack  = false
	DefaultEnableDevMode      = false
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

	rootCmd.AddCommand(EnvCommand())
	rootCmd.AddCommand(DepsCommand())
	rootCmd.AddCommand(LogsCommand())
	rootCmd.AddCommand(InstallCommand())

	rootCmd.PersistentFlags().StringP("context", "", "", "set the Kubernetes context")
	return rootCmd
}

func DepsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   DepsUsage,
		Short: DepsShortDesc,
		Long:  DepsLongDesc,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.AddCommand(depscmd.NewApply().Command())
	cmd.AddCommand(depscmd.NewDelete().Command())

	return cmd
}

func EnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   EnvUsage,
		Short: EnvShortDesc,
		Long:  EnvLongDesc,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.AddCommand(envcmd.NewSync().Command())
	cmd.AddCommand(envcmd.NewClean().Command())
	return cmd
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
	return cmd
}
