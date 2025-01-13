package seaway

import (
	"ctx.sh/seaway/pkg/gen/seaway/v1beta1/seawayv1beta1connect"
	"ctx.sh/seaway/pkg/tracker"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:skip
type Options struct {
	Client        client.Client
	Namespace     string
	StorageURL    string
	StorageBucket string
	StoragePrefix string
	StorageRegion string
	Tracker       *tracker.Tracker
}

// +kubebuilder:skip
type Service struct {
	options *Options
	// TODO: Mutex for uploading artifacts...
}

func RegisterWithWebhook(wh webhook.Server, opts *Options) error {
	service := &Service{
		options: opts,
	}

	path, handler := seawayv1beta1connect.NewSeawayServiceHandler(service)
	wh.Register(path, handler)

	return nil
}
