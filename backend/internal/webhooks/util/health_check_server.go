package util

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
)

type HealthCheckServer struct {
	server *http.Server
}

func NewHealthCheckServer(port int) *HealthCheckServer {
	return &HealthCheckServer{
		server: &http.Server{
			Addr: ":" + strconv.Itoa(port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		},
	}
}

func (hc *HealthCheckServer) Start(ctx context.Context) {
	if err := hc.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Ctx(ctx).Fatal().Err(err).Msg("Health checks HTTP server failed")
	}
}

func (hc *HealthCheckServer) Stop(ctx context.Context) {
	if err := hc.server.Shutdown(ctx); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("Health checks HTTP server shutdown failed")
	}
}
