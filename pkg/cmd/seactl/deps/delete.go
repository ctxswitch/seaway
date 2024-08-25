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

package deps

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	DeleteUsage     = "delete [context]"
	DeleteShortDesc = "Delete the dependencies to the target object storage using the configuration context"
	DeleteLongDesc  = `Delete the dependencies to the target object storage based on the configuration context`
)

type Delete struct {
	logLevel int8
}

func NewDelete() *Delete {
	return &Delete{}
}

func (d *Delete) RunE(cmd *cobra.Command, args []string) error {
	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return nil
}

func (d *Delete) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   DeleteUsage,
		Short: DeleteShortDesc,
		Long:  DeleteLongDesc,
		RunE:  d.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&d.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")

	return cmd
}
