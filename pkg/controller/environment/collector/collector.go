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
	Client                client.Client
	Scheme                *runtime.Scheme
	RegistryURL           string
	RegistryNodePort      uint32
	StorageURL            string
	StorageBucket         string
	StoragePrefix         string
	StorageRegion         string
	StorageForcePathStyle bool
}

func (sc *StateCollector) ObserveAndBuild(ctx context.Context, req ctrl.Request, c *Collection) error {
	observed := NewObservedState()
	observer := &StateObserver{
		Client:  sc.Client,
		Request: req,
	}

	err := observer.observe(ctx, observed)
	if err != nil {
		return err
	}

	c.Observed = observed

	desired := NewDesiredState()
	build := &Builder{
		observed:              observed,
		scheme:                sc.Scheme,
		registryURL:           sc.RegistryURL,
		registryNodePort:      sc.RegistryNodePort,
		storageURL:            sc.StorageURL,
		storageBucket:         sc.StorageBucket,
		storagePrefix:         sc.StoragePrefix,
		storageRegion:         sc.StorageRegion,
		storageForcePathStyle: sc.StorageForcePathStyle,
	}
	err = build.desired(desired)
	if err != nil {
		return err
	}

	c.Desired = desired

	return nil
}
