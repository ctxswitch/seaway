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
