package main

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/arikkfir/devbot/backend/internal/webhooks"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
)

var cfg webhooks.WebhookConfig

func init() {
	initialization.Initialize(&cfg)
}

func main() {
	ctx := context.Background()

	// Setup health check
	hc := util.NewHealthCheckServer(cfg.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Register used CRDs
	err := apiv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register CRDs")
	}

	// Create Kubernetes config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler, err := webhooks.NewPushHandler(kubeConfig, cfg.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create push handler")
	}
	if err := handler.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start push handler")
	}
	defer func(handler *webhooks.PushHandler) {
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
		Handler: util.AccessLogMiddleware(false, nil, mux),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
