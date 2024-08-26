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

package main

import "time"

const (
	DefaultCertDir              string = "/etc/webhook/tls"
	DefaultCertName             string = "tls.crt"
	DefaultKeyName              string = "tls.key"
	DefaultClientCAName         string = "ca.crt"
	DefaultEnableLeaderElection bool   = false
	DefaultSkipInsecureVerify   bool   = false
	DefaultLogLevel             int8   = 0
	DefaultNamespace            string = ""

	ConnectionTimeout time.Duration = 30 * time.Second

	// We need the registry wrapper so we can set up the node ports for the registry.  By default,
	// the install manifests will create a default registry called seaway-registry.
	DefaultRegistryURL      string = "http://registry.seaway-system.svc.cluster.local:5000"
	DefaultRegistryNodePort int32  = 31555
)
