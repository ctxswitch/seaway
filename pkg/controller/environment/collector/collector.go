package collector

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collection struct {
	Observed *ObservedState
	Desired  *DesiredState
}

type StateCollector struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (sc *StateCollector) ObserveAndBuild(ctx context.Context, req ctrl.Request, c *Collection) error {
	observed := NewObservedState()
	observer := &StateObserver{
		Client:  sc.Client,
		Request: req,
	}

	if err := observer.observe(ctx, observed); err != nil {
		return err
	}

	c.Observed = observed

	desired := NewDesiredState()
	build := &Builder{
		observed: observed,
		scheme:   sc.Scheme,
	}
	if err := build.desired(desired); err != nil {
		return err
	}

	c.Desired = desired

	return nil
}
