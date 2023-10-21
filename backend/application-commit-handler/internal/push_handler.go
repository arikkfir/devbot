package internal

import (
	"fmt"
	appsv1 "github.com/arikkfir/devbot/backend/applications-controller/api/v1"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"net/http"
)

type PushHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func NewPushHandler(secret string) *PushHandler {
	ph := PushHandler{}

	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup GitHub webhook")
	}
	ph.Handler = func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)
		if err != nil {
			if errors.Is(err, github.ErrEventNotFound) {
				log.Warn().Err(err).Msg("Unexpected event received - webhook configuration needs to be adjusted")
			} else {
				log.Error().Err(err).Msg("Failed to parse GitHub webhook")
			}
			return
		}
		ph.handlePush(payload.(github.PushPayload))
	}
	return &ph
}

func (ph *PushHandler) handlePush(payload github.PushPayload) {
	app := appsv1.Application{}
	fmt.Printf("%v", app)
	// TODO: implement me
}
