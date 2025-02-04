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
	BuildLogsUsage     = "build [envionment]"
	BuildLogsShortDesc = "Stream logs from the builder pods."
	BuildLogsLongDesc  = `Stream logs from the builder pods.`
)

type BuildLogs struct {
	follow     bool
	tail       int64
	timestamps bool
}

func NewBuildLogs() *BuildLogs {
	return &BuildLogs{}
}
func (l *BuildLogs) RunE(cmd *cobra.Command, args []string) error {
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

	streamer, err := client.NewLogStreamer(kubeContext, corev1.PodLogOptions{
		Follow:     l.follow,
		TailLines:  &l.tail,
		Timestamps: l.timestamps,
		Container:  "builder",
	})
	if err != nil {
		console.Fatal(err.Error())
	}

	labels := fmt.Sprintf("app=%s,group=build", manifest.Name)
	err = streamer.PodLogs(ctx, v1beta1.DefaultControllerNamespace, labels)
	if err != nil && err != context.Canceled {
		console.Fatal(err.Error())
	}

	return nil
}

func (l *BuildLogs) Command() *cobra.Command {
	logsCmd := &cobra.Command{
		Use:   BuildLogsUsage,
		Short: BuildLogsShortDesc,
		Long:  BuildLogsLongDesc,
		RunE:  l.RunE,
	}

	logsCmd.Flags().BoolVarP(&l.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	logsCmd.Flags().Int64VarP(&l.tail, "tail", "", 100, "Number of lines to show from the end of the logs.")
	logsCmd.Flags().BoolVarP(&l.timestamps, "timestamps", "", false, "Include timestamps on each line in the log output.")

	return logsCmd
}
