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
	"github.com/spf13/cobra"
	"stvz.io/coral/pkg/build"
)

const (
	RootUsage     = "seactl [COMMAND] [ARG...]"
	RootShortDesc = "Build controller and image sync tool for kubernetes"
	RootLongDesc  = `Coral is a build controller and image sync tool for kubernetes.  It
provides components for watching source repositories for changes and building containers
when changes and conditions are detected.  It also provides a tool for syncrhonizing the
new images to nodes in a cluster based off of node labels bypassing the need for external
registries.`
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
	}

	rootCmd.AddCommand(NewSync().Command())
	rootCmd.AddCommand(NewListener().Command())
	return rootCmd
}
