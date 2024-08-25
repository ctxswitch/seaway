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
