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
