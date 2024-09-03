package environment

// import (
// 	"context"
// 	"net/url"
// 	"path/filepath"
// 	"testing"

// 	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
// 	"ctx.sh/seaway/pkg/mock"
// 	"github.com/stretchr/testify/assert"
// 	"k8s.io/apimachinery/pkg/types"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/log"
// )

// type WrapHandler struct {
// 	T *testing.T
// 	*Handler
// }

// func NewWrapHandler(ctx context.Context, t *testing.T) *WrapHandler {
// 	wh := &WrapHandler{
// 		T: t,
// 	}

// 	h := mock.NewTestHarness()
// 	log.IntoContext(ctx, h.Logger())

// 	fixtures := filepath.Join("..", "..", "..", "fixtures", "environment_tests")

// 	client := mock.NewClient().
// 		WithLogger(h.Logger()).
// 		WithFixtureDirectory(fixtures)

// 	wh.Handler = &Handler{
// 		client:   client,
// 		observed: NewObservedState(),
// 		desired:  NewDesiredState(),
// 		registry: &url.URL{
// 			Scheme: "http",
// 			Host:   "localhost:5000",
// 		},
// 	}

// 	return wh
// }

// func (wh *WrapHandler) observe(ctx context.Context) {
// 	observed := wh.getObservedState(ctx, wh.T, wh.client)
// 	v1beta1.Defaulted(observed.env)

// 	desired := wh.getDesiredState(wh.T, wh.client, observed)

// 	wh.Handler.observed = observed
// 	wh.Handler.desired = desired
// }

// func (wh *WrapHandler) apply(ctx context.Context, fixture string) {
// 	wh.client.(*mock.Client).ApplyFixtureOrDie(fixture)
// 	wh.observe(ctx)
// }

// func (h *WrapHandler) getObservedState(ctx context.Context, t *testing.T, client client.Client) *ObservedState {
// 	observed := NewObservedState()
// 	observer := &StateObserver{
// 		Client: client,
// 		Request: ctrl.Request{
// 			NamespacedName: types.NamespacedName{
// 				Namespace: "default",
// 				Name:      "test",
// 			},
// 		},
// 	}

// 	err := observer.observe(ctx, observed)
// 	assert.NoError(t, err)

// 	return observed
// }

// func (h *WrapHandler) getDesiredState(t *testing.T, client client.Client, observed *ObservedState) *DesiredState {
// 	desired := NewDesiredState()
// 	build := &Builder{
// 		observed: observed,
// 		scheme:   client.Scheme(),
// 		nodePort: 31555,
// 		registry: &url.URL{
// 			Scheme: "http",
// 			Host:   "localhost:5000",
// 		},
// 	}

// 	err := build.desired(desired)
// 	assert.NoError(t, err)

// 	return desired
// }

// func (h *WrapHandler) reset() {
// 	h.client.(*mock.Client).Reset()
// }

// func TestHandler_reconcileBuildImage(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	handler := NewWrapHandler(ctx, t)
// 	defer handler.reset()

// 	handler.apply(ctx, "test_handler_reconcile_build_image.yaml")
// 	assert.NotNil(t, handler.observed.env)

// 	status := v1beta1.EnvironmentStatus{}

// 	stage, err := handler.reconcileBuildImage(ctx, &status)
// 	assert.NoError(t, err)

// 	assert.Equal(t, v1beta1.EnvironmentStageBuildStarted, stage)

// 	var job batchv1.Job
// 	err = handler.client.Get(ctx, types.NamespacedName{
// 		Name:      handler.observed.env.Name + "-build",
// 		Namespace: handler.observed.env.Namespace,
// 	}, &job)
// 	assert.NoError(t, err)

// 	assert.Equal(t, handler.desired.job, &job)

// 	// Reconcile again since we have the job.  Expect
// }
