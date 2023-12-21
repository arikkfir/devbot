package main

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/backend/internal/webhooks/util"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"strconv"
)

var (
	cfg    github.WebhookConfig
	scheme = runtime.NewScheme()
)

func init() {
	configuration.Parse(&cfg)
	logging.Configure(os.Stderr, cfg.DevMode, cfg.LogLevel)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

func main() {
	ctx := context.Background()

	// Setup health check
	hc := webhooksutil.NewHealthCheckServer(cfg.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Create Kubernetes config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler, err := github.NewPushHandler(kubeConfig, scheme, cfg.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create push handler")
	}
	if err := handler.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start push handler")
	}
	defer func(handler *github.PushHandler) {
		err := handler.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close push handler")
		}
	}(handler)

	// Setup routing
	mux := http.NewServeMux()
	mux.HandleFunc("/github/webhook", handler.HandleWebhookRequest)

	// Setup server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: webhooksutil.AccessLogMiddleware(false, nil, mux),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
