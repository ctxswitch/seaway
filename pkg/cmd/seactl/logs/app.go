package logs

import (
	"context"
	"fmt"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	"ctx.sh/seaway/pkg/kube/client"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

const (
	AppLogsUsage     = "app [envionment]"
	AppLogsShortDesc = "Stream logs from the application pods."
	AppLogsLongDesc  = `Stream logs from the application pods.`
)

type AppLogs struct {
	follow     bool
	tail       int64
	timestamps bool
}

func NewAppLogs() *AppLogs {
	return &AppLogs{}
}
func (l *AppLogs) RunE(cmd *cobra.Command, args []string) error {
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()
	ctx := context.Background()

	if len(args) != 1 {
		return fmt.Errorf("expected environment name")
	}

	var manifest v1beta1.Manifest
	err := manifest.Load("manifest.yaml")
	if err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build environment '%s' not found in the manifest", args[0])
	}

	streamer, err := client.NewLogStreamer(kubeContext, corev1.PodLogOptions{
		Follow:     l.follow,
		TailLines:  &l.tail,
		Timestamps: l.timestamps,
		Container:  "app",
	})
	if err != nil {
		console.Fatal(err.Error())
	}

	labels := fmt.Sprintf("app=%s,group=application", manifest.Name)
	err = streamer.PodLogs(ctx, env.Namespace, labels)
	if err != nil && err != context.Canceled {
		console.Fatal(err.Error())
	}

	return nil
}

func (l *AppLogs) Command() *cobra.Command {
	logsCmd := &cobra.Command{
		Use:   AppLogsUsage,
		Short: AppLogsShortDesc,
		Long:  AppLogsLongDesc,
		RunE:  l.RunE,
	}

	logsCmd.Flags().BoolVarP(&l.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	logsCmd.Flags().Int64VarP(&l.tail, "tail", "", 100, "Number of lines to show from the end of the logs.")
	logsCmd.Flags().BoolVarP(&l.timestamps, "timestamps", "", false, "Include timestamps on each line in the log output.")

	return logsCmd
}
