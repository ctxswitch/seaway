package collector

import (
	"context"
	"path/filepath"
	"testing"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type ObserverTestSuite struct {
	client *mock.Client
	suite.Suite
}

func (s *ObserverTestSuite) SetupTest() {
	logger := zap.New(zap.UseDevMode(true), zap.Level(zapcore.Level(-8)))
	log.SetLogger(logger)

	s.client = mock.NewClient().
		WithLogger(logger).
		WithFixtureDirectory(filepath.Join("..", "..", "..", "..", "fixtures"))

	s.client.ApplyFixtureOrDie("shared", "required.yaml")
}

func (s *ObserverTestSuite) TearDownTest() {
	s.client.Reset()
}

func TestObserverTestSuite(t *testing.T) {
	suite.Run(t, new(ObserverTestSuite))
}

func (s *ObserverTestSuite) TestStateObserver_observe() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.client.ApplyFixtureOrDie(
		"controller_environment_collector",
		"test_state_observer_observe_1.yaml",
	)

	observed := NewObservedState()
	observer := &StateObserver{
		Client: s.client,
		Request: ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	err := observer.observe(ctx, observed)
	s.NoError(err)

	s.NotNil(observed.Env)

	s.Nil(observed.Credentials)
	s.Nil(observed.Job)
	s.Nil(observed.Deployment)
	s.Nil(observed.Service)
	s.Nil(observed.Ingress)

	s.client.ApplyFixtureOrDie(
		"controller_environment_collector",
		"test_state_observer_observe_2.yaml",
	)

	s.NoError(err)

	err = observer.observe(ctx, observed)
	s.NoError(err)

	s.NotNil(observed.Env)

	s.NotNil(observed.Credentials)
	s.NotNil(observed.Job)
	s.NotNil(observed.Deployment)
	s.NotNil(observed.Service)
	s.NotNil(observed.Ingress)
}

func (s *ObserverTestSuite) TestStateObserver_observeEnvironment() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.client.ApplyFixtureOrDie(
		"controller_environment_collector",
		"test_state_observer_observe_environment.yaml",
	)

	observer := &StateObserver{
		Client: s.client,
		Request: ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	env, err := observer.observeEnvironment(ctx)
	s.NoError(err)
	s.NotNil(env)

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

	s.Equal(expected.Spec, env.Spec)
}
