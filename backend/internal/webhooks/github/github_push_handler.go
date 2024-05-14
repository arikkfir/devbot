package github

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
)

type PushHandler struct {
	client.Client
	github.Webhook
}

func NewPushHandler(kubeConfig *rest.Config, s *runtime.Scheme, secret string) (*PushHandler, error) {
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, errors.New("failed to create GitHub webhook", err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: s})
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
	} else if namespace, name, err := ph.findRepoForPayload(ctx, pushPayload.Repository.Owner.Login, pushPayload.Repository.Name); err != nil {
		log.Error().Err(err).Msg("Failed looking up GitHubRepository object")
		w.WriteHeader(http.StatusOK)
	} else if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		repo := &apiv1.Repository{}
		if err := ph.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, repo); err != nil {
			return err
		}
		if repo.ObjectMeta.Annotations == nil {
			repo.ObjectMeta.Annotations = map[string]string{}
		}
		repo.ObjectMeta.Annotations["refresh.devbot.com"] = time.Now().String()
		return client.IgnoreNotFound(ph.Update(ctx, repo))
	}); client.IgnoreNotFound(err) != nil {
		log.Error().Err(err).Msg("Failed annotating GitHubRepository object for reconciliation")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (ph *PushHandler) findRepoForPayload(ctx context.Context, repoOwner, repoName string) (namespace string, name string, err error) {
	repositories := apiv1.RepositoryList{}
	if err := ph.List(ctx, &repositories); err != nil {
		return "", "", errors.New("failed to list GitHub repositories", err)
	}

	for _, r := range repositories.Items {
		if r.Spec.GitHub != nil {
			if r.Spec.GitHub.Owner == repoOwner && r.Spec.GitHub.Name == repoName {
				return r.Namespace, r.Name, nil
			}
		}
	}

	return "", "", errors.New("repository object not found")
}
