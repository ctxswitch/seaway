package stage

import (
	"context"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
)

type Error struct {
	status v1beta1.EnvironmentStage
}

func NewError(s v1beta1.EnvironmentStage) *Error {
	return &Error{
		status: s,
	}
}

func (e *Error) Do(ctx context.Context, status *v1beta1.EnvironmentStatus) (v1beta1.EnvironmentStage, error) {
	return e.status, nil
}

var _ Stage = &Error{}
