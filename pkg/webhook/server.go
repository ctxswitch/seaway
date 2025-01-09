package webhook

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const (
	DefaultPort = 9443
)

type Options struct {
	Host         string
	Port         int
	CertDir      string
	CertName     string
	KeyName      string
	ClientCAName string
	TLSOpts      []func(*tls.Config)
	WebhookMux   *http.ServeMux
}

// NewServer constructs a new webhook.Server from the provided options.  This is
// a 1-1 implemetaion of the webhook.Server interface with the exception of the
// utilizing an http2 server for our webhook server.  This was done to allow for
// us to utilize the same server implementation for both the standard admission
// webhooks and the connect services.  In time, I may add a new command to isolate
// the connect services to their own server (but still utilize the same interface).
// What I don't quite know yet, is whether I want to re-architect the "webhook" server
// into something more generic.  For now I like this implementation because it works
// seamlessly with the existing controller-runtime manager.
// See: https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/webhook/server.go
func NewServer(o Options) webhook.Server {
	return &DefaultServer{
		Options: o,
	}
}

// DefaultServer is the default implementation used for Server.
type DefaultServer struct {
	Options        Options
	webhooks       map[string]http.Handler
	defaultingOnce sync.Once
	started        bool
	mu             sync.Mutex
	webhookMux     *http.ServeMux
}

// setDefaults does defaulting for the Server.
func (o *Options) setDefaults() {
	if o.WebhookMux == nil {
		o.WebhookMux = http.NewServeMux()
	}

	if o.Port <= 0 {
		o.Port = DefaultPort
	}

	if len(o.CertDir) == 0 {
		o.CertDir = filepath.Join(os.TempDir(), "k8s-webhook-server", "serving-certs")
	}

	if len(o.CertName) == 0 {
		o.CertName = "tls.crt"
	}

	if len(o.KeyName) == 0 {
		o.KeyName = "tls.key"
	}
}

func (s *DefaultServer) setDefaults() {
	s.webhooks = map[string]http.Handler{}
	s.Options.setDefaults()
	s.webhookMux = s.Options.WebhookMux
}

// NeedLeaderElection implements the LeaderElectionRunnable interface, which indicates
// the webhook server doesn't need leader election.
func (*DefaultServer) NeedLeaderElection() bool {
	return false
}

// Register marks the given webhook as being served at the given path.
func (s *DefaultServer) Register(path string, hook http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.defaultingOnce.Do(s.setDefaults)
	if _, found := s.webhooks[path]; found {
		panic(fmt.Errorf("can't register duplicate path: %v", path))
	}
	s.webhooks[path] = hook
	// We don't have the instrumentation implementation from controller-runtime, but we may
	// want to see if there's a way to add it later.
	// s.webhookMux.Handle(path, metrics.InstrumentedHook(path, hook))
	s.webhookMux.Handle(path, hook)

	regLog := logger.WithValues("path", path)
	regLog.Info("Registering webhook")
}

// Start runs the server.
// It will install the webhook related resources depend on the server configuration.
func (s *DefaultServer) Start(ctx context.Context) error {
	s.defaultingOnce.Do(s.setDefaults)

	logger.Info("Starting webhook server")

	cfg := &tls.Config{ //nolint:gosec
		NextProtos:         []string{"h2"},
		InsecureSkipVerify: true, //nolint:gosec
	}

	for _, op := range s.Options.TLSOpts {
		op(cfg)
	}

	if cfg.GetCertificate == nil {
		certPath := filepath.Join(s.Options.CertDir, s.Options.CertName)
		keyPath := filepath.Join(s.Options.CertDir, s.Options.KeyName)
		certWatcher, err := certwatcher.New(certPath, keyPath)
		if err != nil {
			return err
		}
		cfg.GetCertificate = certWatcher.GetCertificate

		go func() {
			if err := certWatcher.Start(ctx); err != nil {
				logger.Error(err, "certificate watcher error")
			}
		}()
	}

	if s.Options.ClientCAName != "" {
		certPool := x509.NewCertPool()
		clientCABytes, err := os.ReadFile(filepath.Join(s.Options.CertDir, s.Options.ClientCAName))
		if err != nil {
			return fmt.Errorf("failed to read client CA cert: %w", err)
		}

		ok := certPool.AppendCertsFromPEM(clientCABytes)
		if !ok {
			return fmt.Errorf("failed to append client CA cert to CA pool")
		}

		cfg.ClientCAs = certPool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	listener, err := tls.Listen("tcp", net.JoinHostPort(s.Options.Host, strconv.Itoa(s.Options.Port)), cfg)
	if err != nil {
		return err
	}

	logger.Info("Serving webhook server", "host", s.Options.Host, "port", s.Options.Port)

	srv := &http.Server{
		Addr:              ":9443", // TODO: make this configurable
		Handler:           h2c.NewHandler(s.webhookMux, &http2.Server{}),
		MaxHeaderBytes:    1 << 20,
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down webhook server with timeout of 1 minute")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout
			logger.Error(err, "error shutting down the HTTP server")
		}
		close(idleConnsClosed)
	}()

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}

// StartedChecker returns a healthz.Checker which is healthy after the
// server has been started.
func (s *DefaultServer) StartedChecker() healthz.Checker {
	// Used to connect to the controllers own webhook port.
	config := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
	}

	return func(req *http.Request) error {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.started {
			return fmt.Errorf("webhook server has not been started yet")
		}

		logger.Info("Checking if webhook server is reachable", "request", req.URL.String())

		d := &net.Dialer{Timeout: 10 * time.Second}
		conn, err := tls.DialWithDialer(d, "tcp", net.JoinHostPort(s.Options.Host, strconv.Itoa(s.Options.Port)), config)
		if err != nil {
			return fmt.Errorf("webhook server is not reachable: %w", err)
		}

		if err := conn.Close(); err != nil {
			return fmt.Errorf("webhook server is not reachable: closing connection: %w", err)
		}

		return nil
	}
}

// WebhookMux returns the servers WebhookMux.
func (s *DefaultServer) WebhookMux() *http.ServeMux {
	return s.webhookMux
}

var _ webhook.Server = &DefaultServer{}
