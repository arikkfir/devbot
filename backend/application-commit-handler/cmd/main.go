package main

import (
	"context"
	"github.com/arikkfir/devbot/backend/application-commit-handler/internal"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"net/http"
	"strconv"
)

func main() {
	ctx := context.Background()

	// Setup health check
	hc := internal.NewHealthCheckServer(cfg.HTTP.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Setup push handler
	handler := internal.NewPushHandler(cfg.WebhookSecret)

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
