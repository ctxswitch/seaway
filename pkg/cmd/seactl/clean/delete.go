package clean

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	kube "ctx.sh/seaway/pkg/kube/client"
)

func delete(ctx context.Context, client *kube.KubectlCmd, env v1beta1.ManifestEnvironmentSpec) error {
	// TODO: make the dependency clean func.
	return nil
}
