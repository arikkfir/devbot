package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type PushHandler struct {
	webhook *github.Webhook
	k       client.Client
	r       *redis.Client
	pubSub  *redis.PubSub
}

// NewPushHandler creates a new GitHub Push event handler instance.
//
// This object, when started (calling the [PushHandler.Start] method), will handle GitHub "push" events, and forward
// them to an internal Redis-based Pub/Sub channel for asynchronous processing by the [PushHandler.handlePubSubMessages]
// method.
//
// It should also be stopped by a call to the [PushHandler.Close] method.
func NewPushHandler(kubeConfig *rest.Config, redisClient *redis.Client, secret string) (*PushHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, errors.New("failed to create GitHub webhook", err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, errors.New("failed to create Kubernetes client", err)
	}

	ph := PushHandler{
		webhook: hook,
		k:       k8sClient,
		r:       redisClient,
	}

	return &ph, nil
}

// Start will make this object subscribe to the Redis-based Pub/Sub channel for GitHub "push" events, sent there by the
// [PushHandler.HandleWebhookRequest]; that method should be used as a [http.HandlerFunc] by an HTTP server maintained
// by the user of this object (e.g. the "main" function).
func (ph *PushHandler) Start(ctx context.Context) error {
	ph.pubSub = ph.r.Subscribe(ctx, "github.push")
	go ph.handlePubSubMessages(ctx)
	return nil
}

// Close will stop this object from accepting GitHub "push" events, and will unsubscribe from the Redis Pub/Sub channel.
func (ph *PushHandler) Close() error {
	if err := ph.pubSub.Close(); err != nil {
		return errors.New("failed to close PubSub receiver", err)
	}
	return nil
}

// HandleWebhookRequest is a [http.HandlerFunc] which should be used by an HTTP server maintained by the user of this
// object (e.g. the "main" function).
//
// It will accept & validate GitHub webhook calls of the "push" event type, and if
// valid, forward them to the Redis Pub/Sub channel for asynchronous processing by the
// [PushHandler.handlePubSubMessages] method.
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

// handlePubSubMessages is a blocking method which will receive GitHub "push" events from the Redis Pub/Sub channel, and
// will handle them by updating the status of the relevant Application CRD objects in the Kubernetes cluster.
func (ph *PushHandler) handlePubSubMessages(ctx context.Context) {
	channel := ph.pubSub.Channel(redis.WithChannelSize(10))
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

// handlePushPayload will handle a single GitHub "push" event, by updating the status of the relevant Application CRD
// object in the Kubernetes cluster.
func (ph *PushHandler) handlePushPayload(ctx context.Context, payload github.PushPayload) error {
	apps := apiv1.ApplicationList{}
	if err := ph.k.List(ctx, &apps); err != nil {
		return errors.New("failed to list applications", err)
	}
	for _, app := range apps.Items {
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
	// TODO: ensure that if this fails, we retry (can be solved by a Pub/Sub retry mechanism as long as order is maintained)
	if err := ph.k.Status().Update(ctx, app); err != nil {
		return errors.New("failed to update application status", err)
	}
	return nil
}
