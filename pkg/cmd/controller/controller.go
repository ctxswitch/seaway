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

import (
	"crypto/tls"
	"os"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/controller"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	ControllerUsage     = "controller [ARG...]"
	ControllerShortDesc = "Starts the seaway controller."
	ControllerLongDesc  = `Starts the seaway controller providing management of the seaway resources and services in kubernetes.`
)

type Controller struct {
	certs              string
	certName           string
	keyName            string
	clientCAName       string
	leaderElection     bool
	skipInsecureVerify bool
	logLevel           int8
	namespace          string

	// Registry flags.
	registryURL      string
	registryNodePort int32
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) RunE(cmd *cobra.Command, args []string) error {
	scheme := runtime.NewScheme()

	_ = v1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)

	// TODO: more configurations.
	log := zap.New(
		zap.Level(zapcore.Level(c.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	hookServer := webhook.NewServer(webhook.Options{
		Port:    9443,
		CertDir: c.certs,
		// Weird.  One of these causes an error about 'client didn't provide a certificate'
		// Look at these settings in more detail later.
		// CertName:     DefaultCertName,
		// KeyName:      DefaultKeyName,
		// ClientCAName: DefaultClientCAName,
		TLSOpts: []func(*tls.Config){
			func(config *tls.Config) {
				config.InsecureSkipVerify = c.skipInsecureVerify
			},
		},
	})

	// Register client endpoints for job interactions.
	// hookServer.Register("/status", &client.StatusHandler{})
	// TODO: add more endpoints.

	// Initialize the manager.
	log.Info("initializing manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   c.leaderElection,
		LeaderElectionID: "seaway-leader-lock",
		WebhookServer:    hookServer,
	})

	if err != nil {
		log.Error(err, "unable to initialize manager")
		os.Exit(1)
	}

	// // Register webhooks.
	// if err = (&v1beta1.Registry{}).SetupWebhookWithManager(mgr); err != nil {
	// 	log.Error(err, "unable to create webhook", "webhook", "Registry")
	// 	os.Exit(1)
	// }

	if err = (&v1beta1.Environment{}).SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Environment")
		os.Exit(1)
	}

	// Register controllers.
	if err = controller.SetupWithManager(mgr, &controller.Options{
		RegistryURL:      c.registryURL,
		RegistryNodePort: c.registryNodePort,
	}); err != nil {
		log.Error(err, "unable to setup seaway controllers")
		os.Exit(1)
	}

	// Start the manager process
	log.Info("starting manager")
	err = mgr.Start(ctx)

	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	return err
}

func (c *Controller) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ControllerUsage,
		Short: ControllerShortDesc,
		Long:  ControllerLongDesc,
		RunE:  c.RunE,
	}

	cmd.PersistentFlags().StringVarP(&c.certs, "certs", "", DefaultCertDir, "specify the webhooks certs directory")
	cmd.PersistentFlags().StringVarP(&c.certName, "cert-name", "", DefaultCertName, "specify the webhooks cert name")
	cmd.PersistentFlags().StringVarP(&c.keyName, "key-name", "", DefaultKeyName, "specify the webhooks key name")
	cmd.PersistentFlags().StringVarP(&c.clientCAName, "ca-name", "", DefaultClientCAName, "specify the webhooks client ca name")
	cmd.PersistentFlags().BoolVarP(&c.leaderElection, "enable-leader-election", "", DefaultEnableLeaderElection, "enable leader election")
	cmd.PersistentFlags().BoolVarP(&c.skipInsecureVerify, "skip-insecure-verify", "", DefaultSkipInsecureVerify, "skip certificate verification for the webhooks")
	cmd.PersistentFlags().Int8VarP(&c.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")
	cmd.PersistentFlags().StringVarP(&c.namespace, "namespace", "", DefaultNamespace, "limit the controller to a specific namespace")

	cmd.PersistentFlags().StringVarP(&c.registryURL, "registry-url", "", DefaultRegistryURL, "specify the url to the local registry")
	cmd.PersistentFlags().Int32VarP(&c.registryNodePort, "registry-nodeport", "", DefaultRegistryNodePort, "specify the nodeport for the registry service")
	// TODO: in the future, add flags for registry credentials.
	// One thing that I'll note here is that external repos are not supported, but definitely are not out of the question if there is a proxy in place
	// that can be accessed through the nodeport.  Maybe something as simple as a lightweight nginx container.  I might add something like this later.
	return cmd
}