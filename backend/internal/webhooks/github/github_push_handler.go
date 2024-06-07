package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"io"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"time"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
)

const (
	refreshAnnotationName = "refresh.devbot.com"
)

var (
	ErrRepositoryNotFound       = fmt.Errorf("payload repository not found")
	ErrNotGitHubRepository      = fmt.Errorf("repository not configured for GitHub")
	ErrWebhookConfigMissing     = fmt.Errorf("webhook configuration missing")
	ErrWebhookSecretNameNotSet  = fmt.Errorf("webhook secret name not set")
	ErrWebhookSecretKeyNotSet   = fmt.Errorf("webhook secret key not set")
	ErrWebhookSecretNotFound    = fmt.Errorf("webhook secret not found")
	ErrWebhookSecretKeyNotFound = fmt.Errorf("webhook secret key not found in secret")
	ErrWebhookSecretIsEmpty     = fmt.Errorf("webhook secret is empty")
)

type PushHandler struct {
	client.Client
	github.Webhook
}

func NewPushHandler(kubeConfig *rest.Config, s *runtime.Scheme) (*PushHandler, error) {
	hook, err := github.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub webhook: %w", err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: s})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
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
	l := log.Ctx(ctx)

	// Replace request body with a Tee reader that also writes to a local buffer
	// Restore request's original body reader upon exit (so server can close it)
	origReqBody := r.Body
	defer func() { r.Body = origReqBody }()
	payloadBody := &bytes.Buffer{}
	r.Body = io.NopCloser(io.TeeReader(r.Body, payloadBody))

	// Parse the payload
	payload, err := ph.Parse(r, github.PushEvent, github.PingEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) {
			log.Warn().Err(err).Msg("Unexpected event received - webhook configuration needs to be adjusted")
			w.WriteHeader(http.StatusNotImplemented)
		} else {
			log.Error().Err(err).Msg("Failed to parse GitHub webhook")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Obtain repository owner and name from payload
	var ghRepoOwner, ghRepoName string
	if pingPayload, ok := payload.(github.PingPayload); ok {
		ghRepoOwner, ghRepoName = pingPayload.Repository.Owner.Login, pingPayload.Repository.Name
	} else if pushPayload, ok := payload.(github.PushPayload); ok {
		ghRepoOwner, ghRepoName = pushPayload.Repository.Owner.Login, pushPayload.Repository.Name
	} else {
		log.Error().Type("payload", payload).Msg("Unsupported webhook event")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	l = lang.Ptr(l.With().Str("ghRepoOwner", ghRepoOwner).Str("ghRepoName", ghRepoName).Logger())

	// Find corresponding Repository object
	repo, err := ph.getRepositoryObjectForPayload(ctx, ghRepoOwner, ghRepoName)
	if err != nil {
		if errors.Is(err, ErrRepositoryNotFound) {
			l.Warn().Msg("Repository not found")
			w.WriteHeader(http.StatusNotFound)
		} else {
			l.Error().Err(err).Msg("Failed finding repository object")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	l = lang.Ptr(l.With().Str("k8sRepoNamespace", repo.Namespace).Str("k8sRepoName", repo.Name).Logger())

	// Fetch webhook secret this repository references
	webhookSecret, err := ph.getRepositoryWebhookSecret(ctx, repo)
	if err != nil {
		l.Error().Err(err).Msg("Failed getting webhook secret")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Validate webhook contents with our secret
	signature := r.Header.Get("X-Hub-Signature")
	if len(signature) == 0 {
		l.Error().Msg("Empty or missing signature header 'X-Hub-Signature'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mac := hmac.New(sha1.New, []byte(webhookSecret))
	_, _ = mac.Write(payloadBody.Bytes())
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	actualMAC := []byte(signature[5:])
	if !hmac.Equal(actualMAC, []byte(expectedMAC)) {
		l.Error().Bytes("actualMAC", actualMAC).Bytes("expectedMAC", []byte(expectedMAC)).Msg("Failed verifying payload signature")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Detect event type and act
	if _, ok := payload.(github.PingPayload); ok {
		log.Info().Msg("Received ping event from GitHub")
		w.WriteHeader(http.StatusOK)
		return
	} else if _, ok := payload.(github.PushPayload); ok {
		f := func() error { return ph.annotateRepository(ctx, repo.Namespace, repo.Name) }
		if err := retry.RetryOnConflict(retry.DefaultBackoff, f); err != nil {
			log.Error().Err(err).Msg("Failed annotating repository for reconciliation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		log.Error().Type("payload", payload).Msg("Unsupported webhook event")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	// TODO: fail webhook call if repository is in Invalid condition due to webhook configuration

	w.WriteHeader(http.StatusOK)
}

func (ph *PushHandler) getRepositoryObjectForPayload(ctx context.Context, owner, name string) (*apiv1.Repository, error) {
	repositories := apiv1.RepositoryList{}
	if err := ph.List(ctx, &repositories); err != nil {
		return nil, fmt.Errorf("failed to list GitHub repositories: %w", err)
	}

	for _, r := range repositories.Items {
		if r.Spec.GitHub != nil {
			if r.Spec.GitHub.Owner == owner && r.Spec.GitHub.Name == name {
				return &r, nil
			}
		}
	}

	return nil, ErrRepositoryNotFound
}

func (ph *PushHandler) getRepositoryWebhookSecret(ctx context.Context, repo *apiv1.Repository) (string, error) {

	// Validate repository is indeed a GitHub repository
	if repo.Spec.GitHub == nil {
		return "", ErrNotGitHubRepository
	} else if repo.Spec.GitHub.WebhookSecret == nil {
		return "", ErrWebhookConfigMissing
	}
	webhookSecretCfg := repo.Spec.GitHub.WebhookSecret

	// Validate auth secret name & key are not missing
	if webhookSecretCfg.Secret.Name == "" {
		return "", ErrWebhookSecretNameNotSet
	} else if webhookSecretCfg.Key == "" {
		return "", ErrWebhookSecretKeyNotSet
	}

	// Fetch secret
	webhookSecret := &v12.Secret{}
	secretObjKey := webhookSecretCfg.Secret.GetObjectKey(repo.Namespace)
	if err := ph.Client.Get(ctx, secretObjKey, webhookSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", ErrWebhookSecretNotFound
		} else if apierrors.IsForbidden(err) {
			return "", fmt.Errorf("webhook secret is forbidden: %w", err)
		} else {
			return "", fmt.Errorf("webhook secret could not be read: %w", err)
		}
	}

	// Extract webhook secret
	secretValue, ok := webhookSecret.Data[webhookSecretCfg.Key]
	if !ok {
		return "", ErrWebhookSecretKeyNotFound
	} else if string(secretValue) == "" {
		return "", ErrWebhookSecretIsEmpty
	}

	return string(secretValue), nil
}

func (ph *PushHandler) annotateRepository(ctx context.Context, namespace, name string) error {
	repo := &apiv1.Repository{}
	if err := ph.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, repo); err != nil {
		return err
	}

	if repo.ObjectMeta.Annotations == nil {
		repo.ObjectMeta.Annotations = map[string]string{}
	}

	repo.ObjectMeta.Annotations[refreshAnnotationName] = time.Now().String()
	return client.IgnoreNotFound(ph.Update(ctx, repo))
}
