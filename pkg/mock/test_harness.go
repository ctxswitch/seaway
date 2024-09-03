package mock

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type TestHarness struct {
	logger   logr.Logger
	setup    func(context.Context, *testing.T)
	teardown func(context.Context, *testing.T)
}

func NewTestHarness() *TestHarness {
	logger := zap.New(zap.UseDevMode(true), zap.Level(zapcore.Level(-8)))
	// Use the logger for the controller-runtime components as well.
	logf.SetLogger(logger)

	return &TestHarness{
		logger:   logger,
		setup:    func(ctx context.Context, t *testing.T) {},
		teardown: func(ctx context.Context, t *testing.T) {},
	}
}

func (h *TestHarness) Setup(fn func(context.Context, *testing.T)) {
	h.setup = fn
}

func (h *TestHarness) TearDown(fn func(context.Context, *testing.T)) {
	h.teardown = fn
}

func (h *TestHarness) Run(ctx context.Context, t *testing.T, funcs ...func(ctx context.Context, t *testing.T)) {
	for _, f := range funcs {
		h.setup(ctx, t)
		f(ctx, t)
		h.teardown(ctx, t)
	}
}

func (h *TestHarness) Logger() logr.Logger {
	return h.logger
}
