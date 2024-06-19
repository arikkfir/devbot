package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/arikkfir/command"
	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/lang"
	"github.com/arikkfir/devbot/internal/util/observability"
	"github.com/arikkfir/devbot/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/internal/webhooks/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Action struct {
	HealthPort  int `required:"true" desc:"Health endpoint port."`
	MetricsPort int `required:"true" desc:"Metrics endpoint port."`
	ServerPort  int `required:"true" desc:"Webhook server port."`
}

func (e *Action) Run(ctx context.Context) error {

	// Setup Kubernetes scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))

	// Create Kubernetes config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create the webhook handlers
	handler, err := github.NewPushHandler(kubeConfig, scheme)
	if err != nil {
		return fmt.Errorf("failed to create push handler: %w", err)
	}

	// Setup servers
	servers := []*http.Server{
		e.newHealthCheckServer(),
		e.newMetricsHTTPServer(),
		e.newWebhooksHTTPServer(handler),
	}

	// Start the servers
	stop := make(chan string, 100)
	errs := make(chan error, 100)
	for _, server := range servers {
		go e.startHTTPServer(stop, errs, "health", server)
	}

	// Wait for either one of the HTTP servers to prematurely exit, or an OS interrupt signal
	select {
	case name := <-stop:
		log.Error().Str("server", name).Msg("One of the HTTP servers failed")
	case <-ctx.Done():
		log.Error().Msg("Interrupt signal received")
	}

	// Gracefully shutdown all HTTP servers
	for _, server := range servers {
		err = errors.Join(err, server.Shutdown(context.Background()))
	}

	// Close the errors channel & collect all errors that occurred so far
	close(errs)
	for e := range errs {
		err = errors.Join(err, e)
	}
	return err
}

func (e *Action) startHTTPServer(stopChan chan string, errChan chan error, name string, server *http.Server) {
	if err := server.ListenAndServe(); lang.IgnoreErrorOfType(err, http.ErrServerClosed) != nil {
		errChan <- fmt.Errorf("%s server failed: %w", name, err)
		stopChan <- name
	}
}

func (e *Action) newHealthCheckServer() *http.Server {
	return &http.Server{
		Addr: ":" + strconv.Itoa(e.HealthPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}
}

func (e *Action) newMetricsHTTPServer() *http.Server {
	return &http.Server{
		Addr:    ":" + strconv.Itoa(e.MetricsPort),
		Handler: promhttp.Handler(),
	}
}

func (e *Action) newWebhooksHTTPServer(handler *github.PushHandler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/github/webhook", handler.HandleWebhookRequest)
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(e.ServerPort),
		Handler: webhooksutil.AccessLogMiddleware(false, nil, mux),
	}
	return server
}

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Devbot GitHub webhook connects GitHub events to Devbot installations.",
		`This webhook will receive events from GitHub and mark the corresponding GitHub repository accordingly.'`,
		&Action{
			HealthPort:  9000,
			MetricsPort: 8000,
			ServerPort:  8080,
		},
		[]any{
			&observability.LoggingHook{LogLevel: "info"},
			&observability.OTelHook{
				ServiceName:         "devbot-webhooks",
				MetricReaderFactory: func(ctx context.Context) (metric.Reader, error) { return prometheus.New() },
			},
		},
	)

	os.Exit(int(command.Execute(os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))
}
