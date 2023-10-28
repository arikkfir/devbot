package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"strings"
)

var (
	GetApplicationsURI = fmt.Sprintf("/apis/%s/%s/applications", apiv1.GroupVersion.Group, apiv1.GroupVersion.Version)
)

type PushHandler struct {
	webhook     *github.Webhook
	k8sClient   *kubernetes.Clientset
	redisClient *redis.Client
	pubsub      *redis.PubSub
}

func NewPushHandler(kubeConfig *rest.Config, redisClient *redis.Client, secret string) (*PushHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, errors.New("failed to create GitHub webhook", err)
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.New("failed to create Kubernetes client", err)
	}

	ph := PushHandler{
		webhook:     hook,
		k8sClient:   k8sClient,
		redisClient: redisClient,
	}

	return &ph, nil
}

func (ph *PushHandler) Start(ctx context.Context) error {
	ph.pubsub = ph.redisClient.Subscribe(ctx, "github.push")
	go ph.handlePubsubMessages()
	return nil
}

func (ph *PushHandler) Close() error {
	if err := ph.pubsub.Close(); err != nil {
		return errors.New("failed to close PubSub receiver", err)
	}
	return nil
}

func (ph *PushHandler) handlePubsubMessages() {
	channel := ph.pubsub.Channel(redis.WithChannelSize(10))
	for {
		// TODO: verify correct pub/sub semantics (ack/nack, dead-letter, etc)

		msg, ok := <-channel
		if !ok {
			log.Info().Msg("Closing push-handler PubSub receiver")
			return
		}

		decoder := json.NewDecoder(strings.NewReader(msg.Payload))
		var payload github.PushPayload
		if err := decoder.Decode(&payload); err != nil {
			log.Error().Err(err).Str("payload", msg.Payload).Msg("Failed to decode payload")
		} else if err := ph.handlePushEvent(context.Background(), payload); err != nil {
			log.Error().Err(err).Str("payload", msg.Payload).Msg("Failed to handle push event")
		}
	}
}

func (ph *PushHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	payload, err := ph.webhook.Parse(r, github.PushEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) {
			log.Warn().Err(err).Msg("Unexpected event received - webhook configuration needs to be adjusted")
		} else {
			log.Error().Err(err).Msg("Failed to parse GitHub webhook")
		}
		return
	}

	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(payload); err != nil {
		log.Error().Err(err).Msg("Failed to encode payload")
		w.WriteHeader(http.StatusInternalServerError)
	} else if err := ph.redisClient.Publish(r.Context(), "github.push", buf.Bytes()).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish payload to Redis")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (ph *PushHandler) handlePushEvent(ctx context.Context, payload github.PushPayload) error {
	raw, err := ph.k8sClient.RESTClient().Get().AbsPath(GetApplicationsURI).DoRaw(ctx)
	if err != nil {
		return errors.New("failed to get applications", err)
	}

	applications := apiv1.ApplicationList{}
	if err := json.Unmarshal(raw, &applications); err != nil {
		return errors.New("failed to unmarshal applications", err)
	}

	for _, app := range applications.Items {
		if app.Spec.Repository.GitHub == nil {
			continue
		} else if app.Spec.Repository.GitHub.Owner != payload.Repository.Owner.Login {
			continue
		} else if app.Spec.Repository.GitHub.Name != payload.Repository.Name {
			continue
		} else if err := ph.handleApplication(ctx, payload, app); err != nil {
			return errors.New("failed to handle push event for application '%s'", app.Name, err)
		} else {
			return nil
		}
	}
	return errors.New("application not found for push event", errors.Meta("payload", payload))
}

func (ph *PushHandler) handleApplication(_ context.Context, payload github.PushPayload, app apiv1.Application) error {
	if payload.Deleted {
		if app.Status.Refs != nil {
			delete(app.Status.Refs, payload.Ref)
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
		refStatus.LatestAvailableCommit = payload.After
	}
	return nil
}
