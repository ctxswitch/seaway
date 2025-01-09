package healthz

import (
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type Options struct{}

type Service struct {
	options *Options
}

func RegisterWithWebhook(wh webhook.Server, opts *Options) error {
	service := &Service{
		options: opts,
	}

	wh.Register("/healthz", service)
	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
