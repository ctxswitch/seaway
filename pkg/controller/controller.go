// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"ctx.sh/seaway/pkg/controller/environment"
	"ctx.sh/seaway/pkg/tracker"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	RegistryURL           string
	RegistryNodePort      uint32
	StorageURL            string
	StorageBucket         string
	StoragePrefix         string
	StorageRegion         string
	StorageForcePathStyle bool
	Tracker               *tracker.Tracker
}

type Controller struct{}

// SetupWithManager sets up any known controllers.
func SetupWithManager(mgr ctrl.Manager, opts *Options) error {
	return environment.SetupWithManager(mgr, &environment.Options{
		RegistryURL:           opts.RegistryURL,
		RegistryNodePort:      opts.RegistryNodePort,
		StorageURL:            opts.StorageURL,
		StorageBucket:         opts.StorageBucket,
		StoragePrefix:         opts.StoragePrefix,
		StorageRegion:         opts.StorageRegion,
		StorageForcePathStyle: opts.StorageForcePathStyle,
		Tracker:               opts.Tracker,
	})
}
