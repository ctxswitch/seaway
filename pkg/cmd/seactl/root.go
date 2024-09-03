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
	initcmd "ctx.sh/seaway/pkg/cmd/seactl/init"
	"github.com/spf13/cobra"
)

const (
	RootUsage     = "seactl [command] [args...]"
	RootShortDesc = "CLI utility for managing Seaway development environments."
	RootLongDesc  = `CLI utility for managing Seaway development environments.`
	// TODO: Make these descriptions more informational.
	DepsUsage     = "deps [subcommand] [context]"
	DepsShortDesc = "Utility to manage application dependencies"
	DepsLongDesc  = `Utility to manage application dependencies`
	EnvUsage      = "env [subcommand] [context]"
	EnvShortDesc  = "Utility to manage development environments"
	EnvLongDesc   = `Utility to manage development environments`
	InitUsage     = "init [subcommand]"
	InitShortDesc = "Utility to initialize Seaway resources"
	InitLongDesc  = `Utility to initialize Seaway resources`
)

type Root struct{}

func NewRoot() *Root {
	return &Root{}
}

func (r *Root) Execute() error {
	if err := r.Command().Execute(); err != nil {
		return err
	}

	return nil
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
	rootCmd.AddCommand(InitCommand())

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

func InitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   InitUsage,
		Short: InitShortDesc,
		Long:  InitLongDesc,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.AddCommand(initcmd.NewShared().Command())
	return cmd
}
