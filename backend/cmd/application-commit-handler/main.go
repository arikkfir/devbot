package main

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/arikkfir/devbot/backend/internal/webhooks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
)

var cfg webhooks.WebhookCommandConfig

func init() {
	initialization.Initialize(&cfg)
}

func main() {
	ctx := context.Background()

	// Setup health check
	hc := util.NewHealthCheckServer(cfg.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Setup Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   0,
	})

	// Register used CRDs
	err := apiv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register CRDs")
	}

	// Create Kubernetes client
	k8sClient, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler, err := webhooks.NewPushHandler(k8sClient, redisClient, cfg.Webhook.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create push handler")
	}
	defer func(handler *webhooks.PushHandler) {
		err := handler.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close push handler")
		}
	}(handler)

	// Setup server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Webhook.Port),
		Handler: util.AccessLogMiddleware(false, nil, handler.HandleRequest),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
