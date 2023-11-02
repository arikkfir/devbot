package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"net/http"
	"strings"
)

type PushHandler struct {
	webhook *github.Webhook
	k       *util.K8sClient
	r       *redis.Client
	pubsub  *redis.PubSub
}

func NewPushHandler(k8sClient *util.K8sClient, redisClient *redis.Client, secret string) (*PushHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, errors.New("failed to create GitHub webhook", err)
	}

	ph := PushHandler{
		webhook: hook,
		k:       k8sClient,
		r:       redisClient,
	}

	return &ph, nil
}

func (ph *PushHandler) Start(ctx context.Context) error {
	ph.pubsub = ph.r.Subscribe(ctx, "github.push")
	go ph.handlePubsubMessages(ctx)
	return nil
}

func (ph *PushHandler) Close() error {
	if err := ph.pubsub.Close(); err != nil {
		return errors.New("failed to close PubSub receiver", err)
	}
	return nil
}

func (ph *PushHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	payload, err := ph.webhook.Parse(r, github.PushEvent, github.PingEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) {
			log.Warn().Err(err).Msg("Unexpected event received - webhook configuration needs to be adjusted")
		} else {
			log.Error().Err(err).Msg("Failed to parse GitHub webhook")
		}
		return
	} else if _, ok := payload.(github.PingPayload); ok {
		log.Info().Msg("Received ping event from GitHub")
		w.WriteHeader(http.StatusOK)
		return
	}

	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(payload); err != nil {
		log.Error().Err(err).Msg("Failed to encode payload")
		w.WriteHeader(http.StatusInternalServerError)
	} else if err := ph.r.Publish(r.Context(), "github.push", buf.Bytes()).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish payload to Redis")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (ph *PushHandler) handlePubsubMessages(ctx context.Context) {
	channel := ph.pubsub.Channel(redis.WithChannelSize(10))
	for {
		// TODO: verify correct pub/sub semantics (ack/nack, dead-letter, etc)

		msg, ok := <-channel
		if !ok {
			log.Info().Msg("Closing push-handler PubSub receiver")
			return
		}

		log.Info().Interface("msg", msg).Msg("Received message from PubSub")

		decoder := json.NewDecoder(strings.NewReader(msg.Payload))
		var payload github.PushPayload
		if err := decoder.Decode(&payload); err != nil {
			log.Error().Err(err).Str("payload", msg.Payload).Msg("Failed to decode payload")
		} else if err := ph.handlePushPayload(ctx, payload); err != nil {
			log.Error().Err(err).Str("payload", msg.Payload).Msg("Failed to handle push event")
		}
	}
}

func (ph *PushHandler) handlePushPayload(ctx context.Context, payload github.PushPayload) error {
	applications, err := ph.k.GetApplications(ctx)
	if err != nil {
		return errors.New("failed to get applications", err)
	}

	for _, app := range applications {
		if app.Spec.Repository.GitHub == nil {
			continue
		} else if app.Spec.Repository.GitHub.Owner != payload.Repository.Owner.Login {
			continue
		} else if app.Spec.Repository.GitHub.Name != payload.Repository.Name {
			continue
		} else if err := ph.handleApplication(ctx, payload, &app); err != nil {
			return errors.New("failed to handle push event for application '%s'", app.Name, err)
		} else {
			return nil
		}
	}
	return errors.New("application not found for push event", errors.Meta("payload", payload))
}

func (ph *PushHandler) handleApplication(ctx context.Context, payload github.PushPayload, app *apiv1.Application) error {
	if payload.Deleted {
		if app.Status.Refs != nil {
			delete(app.Status.Refs, payload.Ref)
		} else {
			return nil
		}
	} else {
		if app.Status.Refs == nil {
			app.Status.Refs = make(map[string]apiv1.RefStatus)
		}
		refStatus, ok := app.Status.Refs[payload.Ref]
		if !ok {
			refStatus = apiv1.RefStatus{}
			app.Status.Refs[payload.Ref] = refStatus
		}
		if refStatus.LatestAvailableCommit != payload.After {
			refStatus.LatestAvailableCommit = payload.After
		} else {
			return nil
		}
	}
	if err := ph.k.UpdateApplicationStatus(ctx, app); err != nil {
		return errors.New("failed to update application status", err)
	} else {
		return nil
	}
}
