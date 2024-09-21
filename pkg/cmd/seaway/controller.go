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
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1/handlers"
	"ctx.sh/seaway/pkg/controller"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	defaultConfig      string
}

func NewController() *Controller {
	return &Controller{}
}

// RunE configures and starts the seaway controller.
func (c *Controller) RunE(cmd *cobra.Command, args []string) error {
	scheme := runtime.NewScheme()

	_ = v1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)

	// TODO: more configurations.
	log := zap.New(
		zap.Level(zapcore.Level(c.logLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	hookServer := webhook.NewServer(webhook.Options{
		Port:    9443,
		CertDir: c.certs,
		// TODO: One of these causes an error about 'client didn't provide a certificate'
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

	hookServer.Register("/upload", handlers.NewUploadHandler(&handlers.UploadOptions{
		Client: mgr.GetClient(),
	}))

	hookServer.Register("/ping", handlers.NewPingHandler())

	if err = (&v1beta1.Environment{}).SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Environment")
		os.Exit(1)
	}

	if err = controller.SetupWithManager(mgr, &controller.Options{}); err != nil {
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

// Command returns the cobra command for the controller.
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
	cmd.PersistentFlags().StringVarP(&c.defaultConfig, "default-config", "", DefaultConfigName, "specify the default seaway config that will be used if none is specified")

	return cmd
}
