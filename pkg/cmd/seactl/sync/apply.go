package sync

import (
	"context"
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/cmd/util"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/util/kustomize"
)

func apply(ctx context.Context, client *kube.KubectlCmd, env v1beta1.ManifestEnvironmentSpec) error {
	console.Section("Applying dependencies for the '%s' environment", env.Name)
	for _, dep := range env.Dependencies {
		// TODO: pull the errors back to here.
		_ = applyResource(ctx, client, dep)
	}

	return nil
}

func applyResource(ctx context.Context, client *kube.KubectlCmd, dep v1beta1.ManifestDependency) error {
	krusty, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: dep.Path,
	})
	if err != nil {
		console.Fatal("Unable to initialize kustomize: %s", err.Error())
		return err
	}

	err = krusty.Build()
	if err != nil {
		console.Fatal("Unable to perform kustomize build: %s", err.Error())
		return err
	}

	items := krusty.Resources()

	console.Info("Applying dependency '%s'", dep.Path)
	for _, item := range items {
		if err := util.Apply(ctx, client, item); err != nil {
			console.Fatal("error: %s", err.Error())
		}
	}

	if len(dep.Wait) == 0 {
		return nil
	}

	console.Info("Waiting for resource conditions")
	for _, cond := range dep.Wait {
		if obj, ok := krusty.GetResource(cond.Kind, cond.Name); ok {
			err := util.Wait(ctx, client, obj, cond)
			if err != nil {
				console.Fatal("error: %s", err.Error())
			}
		}
	}

	return nil
}
