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
	"ctx.sh/seaway/pkg/cmd/seaway/operator"
	"github.com/spf13/cobra"
)

const (
	RootUsage     = "seaway [command] [ARG...]"
	RootShortDesc = "Build controller and image sync tool for kubernetes."
	RootLongDesc  = `Build controller and image sync tool for kubernetes.`

	OperatorUsage     = "operator [ARG...]"
	OperatorShortDesc = "Run the seaway operator."
	OperatorLongDesc  = `Run the seaway operator.`
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
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	rootCmd.AddCommand(OperatorCommand())
	return rootCmd
}

// Command returns the cobra command for the controller.
func OperatorCommand() *cobra.Command {
	c := operator.Command{}

	cmd := &cobra.Command{
		Use:   OperatorUsage,
		Short: OperatorShortDesc,
		Long:  OperatorLongDesc,
		RunE:  c.RunE,
	}

	cmd.PersistentFlags().StringVarP(&c.Certs, "certs", "", DefaultCertDir, "specify the webhooks certs directory")
	cmd.PersistentFlags().StringVarP(&c.CertName, "cert-name", "", DefaultCertName, "specify the webhooks cert name")
	cmd.PersistentFlags().StringVarP(&c.KeyName, "key-name", "", DefaultKeyName, "specify the webhooks key name")
	cmd.PersistentFlags().StringVarP(&c.ClientCAName, "ca-name", "", DefaultClientCAName, "specify the webhooks client ca name")
	cmd.PersistentFlags().BoolVarP(&c.LeaderElection, "enable-leader-election", "", DefaultEnableLeaderElection, "enable leader election")
	cmd.PersistentFlags().BoolVarP(&c.SkipInsecureVerify, "skip-insecure-verify", "", DefaultSkipInsecureVerify, "skip certificate verification for the webhooks")
	cmd.PersistentFlags().Int8VarP(&c.LogLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().StringVarP(&c.Namespace, "namespace", "", DefaultNamespace, "limit the controller to a specific namespace")
	cmd.PersistentFlags().StringVarP(&c.DefaultConfig, "default-config", "", DefaultConfigName, "specify the default seaway config that will be used if none is specified")
	cmd.PersistentFlags().StringVarP(&c.BuildNamespace, "build-namespace", "", DefaultBuildNamespace, "specify the namespace that will be used for application builds")
	return cmd
}
