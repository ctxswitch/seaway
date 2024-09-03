package collector

import (
	"context"
	"path/filepath"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestStateObserver_observe(t *testing.T) {
	var client *mock.Client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	client = mock.NewClient().
		WithLogger(h.Logger()).
		WithFixtureDirectory(filepath.Join("..", "..", "..", "..", "fixtures", "controller_environment_collector"))
	defer client.Reset()

	client.ApplyFixtureOrDie("test_state_observer_observe_0.yaml")
	client.ApplyFixtureOrDie("test_state_observer_observe_1.yaml")

	observed := NewObservedState()
	observer := &StateObserver{
		Client: client,
		Request: ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	err := observer.observe(ctx, observed)
	assert.NoError(t, err)

	assert.NotNil(t, observed.Env)
	assert.NotNil(t, observed.UserSecret)

	assert.Nil(t, observed.Job)
	assert.Nil(t, observed.Deployment)
	assert.Nil(t, observed.Service)
	assert.Nil(t, observed.Ingress)

	client.ApplyFixtureOrDie("test_state_observer_observe_2.yaml")
	assert.NoError(t, err)

	err = observer.observe(ctx, observed)
	assert.NoError(t, err)

	assert.NotNil(t, observed.Env)
	assert.NotNil(t, observed.UserSecret)

	assert.NotNil(t, observed.Job)
	assert.NotNil(t, observed.Deployment)
	assert.NotNil(t, observed.Service)
	assert.NotNil(t, observed.Ingress)
}

func TestStateObserver_observeEnvironment(t *testing.T) {
	var client *mock.Client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := mock.NewTestHarness()
	log.IntoContext(ctx, h.Logger())

	client = mock.NewClient().
		WithLogger(h.Logger()).
		WithFixtureDirectory(filepath.Join("..", "..", "..", "..", "fixtures", "controller_environment_collector"))
	defer client.Reset()

	client.ApplyFixtureOrDie("test_state_observer_observe_environment.yaml")

	observer := &StateObserver{
		Client: client,
		Request: ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	env, err := observer.observeEnvironment(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// Ensure that the environment is found and defaulted.
	expected := &v1beta1.Environment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "seaway.ctx.sh/v1beta1",
			Kind:       "Environment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	v1beta1.Defaulted(expected)

	assert.Equal(t, expected.Spec, env.Spec)
}
