package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
)

type Stage interface {
	Do(context.Context, *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error)
}
