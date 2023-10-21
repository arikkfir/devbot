package main

import (
	"context"
	"github.com/arikkfir/devbot/backend/application-commit-handler/internal"
	appsv1 "github.com/arikkfir/devbot/backend/applications-controller/api/v1"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
)

func main() {
	ctx := context.Background()

	// Setup health check
	hc := internal.NewHealthCheckServer(cfg.HTTP.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Register used CRDs
	err := appsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register CRDs")
	}

	// Create Kubernetes client
	k8sClient, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler := internal.NewPushHandler(k8sClient, cfg.WebhookSecret)

	// Setup server
	server := &http.Server{
		Addr: ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler: internal.AccessLogMiddleware(
			cfg.HTTP.AccessLogExcludeRemoteAddr,
			cfg.HTTP.AccessLogExcludedHeaders,
			handler.Handler,
		),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
