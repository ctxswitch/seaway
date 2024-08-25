package controller

import (
	"ctx.sh/seaway/pkg/controller/environment"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	RegistryURL      string
	RegistryNodePort int32
}

type Controller struct{}

func SetupWithManager(mgr ctrl.Manager, opts *Options) (err error) {
	// if err = registry.SetupWithManager(mgr); err != nil {
	// 	return
	// }

	if err = environment.SetupWithManager(mgr, opts.RegistryURL, opts.RegistryNodePort); err != nil {
		return
	}

	return
}
