package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type PushHandler struct {
	client.Client
	github.Webhook
}

func NewPushHandler(kubeConfig *rest.Config, secret string) (*PushHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, errors.New("failed to create GitHub webhook", err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, errors.New("failed to create Kubernetes client", err)
	}

	return &PushHandler{Client: k8sClient, Webhook: *hook}, nil
}

func (ph *PushHandler) Start(_ context.Context) error {
	return nil
}

func (ph *PushHandler) Close() error {
	return nil
}

func (ph *PushHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	payload, err := ph.Webhook.Parse(r, github.PushEvent, github.PingEvent)
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
	} else if pushPayload, ok := payload.(github.PushPayload); !ok {
		log.Error().Type("payloadType", payload).Msg("Unexpected payload type received")
		w.WriteHeader(http.StatusBadRequest)
	} else if repo, err := ph.findRepoForPayload(ctx, pushPayload.Repository.Owner.Login, pushPayload.Repository.Name); err != nil {
		log.Error().Err(err).Msg("Failed looking up GitHubRepository object")
		w.WriteHeader(http.StatusBadRequest)
	} else if err := ph.annotateRepoForReconciliation(ctx, repo); err != nil {
		log.Error().Err(err).Msg("Failed annotating GitHubRepository object for reconciliation")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (ph *PushHandler) findRepoForPayload(ctx context.Context, repoOwner, repoName string) (*apiv1.GitHubRepository, error) {
	repositories := apiv1.GitHubRepositoryList{}
	if err := ph.List(ctx, &repositories); err != nil {
		return nil, errors.New("failed to list GitHub repositories", err)
	}

	for _, r := range repositories.Items {
		if r.Spec.Owner == repoOwner && r.Spec.Name == repoName {
			return &r, nil
		}
	}

	return nil, nil
}

func (ph *PushHandler) annotateRepoForReconciliation(ctx context.Context, repo *apiv1.GitHubRepository) error {
	if repo.ObjectMeta.Annotations == nil {
		repo.ObjectMeta.Annotations = map[string]string{}
	}
	repo.ObjectMeta.Annotations["refresh.devbot.com"] = time.Now().String()
	if err := ph.Update(ctx, repo); err != nil {
		return errors.New("failed annotating GitHubRepository object for reconciliation", err)
	}
	return nil
}
