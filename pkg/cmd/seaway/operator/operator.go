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

package operator

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

type Command struct {
	Certs              string
	CertName           string
	KeyName            string
	ClientCAName       string
	LeaderElection     bool
	SkipInsecureVerify bool
	LogLevel           int8
	Namespace          string
	DefaultConfig      string
	BuildNamespace     string
}

func NewCommand() *Command {
	return &Command{}
}

// RunE configures and starts the seaway controller.
func (c *Command) RunE(cmd *cobra.Command, args []string) error {
	scheme := runtime.NewScheme()

	_ = v1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)

	// TODO: more configurations.
	log := zap.New(
		zap.Level(zapcore.Level(c.LogLevel) * -1),
	)

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(log)

	hookServer := webhook.NewServer(webhook.Options{
		Port:    9443,
		CertDir: c.Certs,
		// TODO: One of these causes an error about 'client didn't provide a certificate'
		// Look at these settings in more detail later.
		// CertName:     DefaultCertName,
		// KeyName:      DefaultKeyName,
		// ClientCAName: DefaultClientCAName,
		TLSOpts: []func(*tls.Config){
			func(config *tls.Config) {
				config.InsecureSkipVerify = c.SkipInsecureVerify
			},
		},
	})

	// Initialize the manager.
	log.Info("initializing manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   c.LeaderElection,
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
