package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
)

type Error struct{}

func NewError() *Error {
	return &Error{}
}

func (e *Error) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	return v1beta1.EnvironmentStageFailed, nil
}

var _ Stage = &Error{}
