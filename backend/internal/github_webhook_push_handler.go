package internal

import (
	"context"
	"encoding/json"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
)

type PushHandler struct {
	k8sClient *rest.Config
	Handler   func(w http.ResponseWriter, r *http.Request)
}

func NewPushHandler(k8sClient *rest.Config, secret string) *PushHandler {
	ph := PushHandler{
		k8sClient: k8sClient,
	}

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

		clientset, err := kubernetes.NewForConfig(ph.k8sClient)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
		}

		ph.handlePush(r.Context(), clientset, payload.(github.PushPayload))
	}
	return &ph
}

func (ph *PushHandler) handlePush(ctx context.Context, clientset *kubernetes.Clientset, payload github.PushPayload) {
	path := fmt.Sprintf("/apis/%s/%s/applications", apiv1.GroupVersion.Group, apiv1.GroupVersion.Version)
	raw, err := clientset.RESTClient().Get().AbsPath(path).DoRaw(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get applications")
		return
	}

	applications := apiv1.ApplicationList{}
	if err := json.Unmarshal(raw, &applications); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal applications")
		return
	}

	for _, app := range applications.Items {
		if app.Spec.Repository.GitHub != nil {
			appRepo := app.Spec.Repository.GitHub
			if appRepo.Owner == payload.Repository.Owner.Login && appRepo.Name == payload.Repository.Name {
				ph.handleApplication(ctx, clientset, payload, app)
				return
			}
		}
	}

	log.Warn().Str("repository", payload.Repository.FullName).Msg("Received push event for an unknown repository")
}

func (ph *PushHandler) handleApplication(_ context.Context, _ *kubernetes.Clientset, payload github.PushPayload, app apiv1.Application) {
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
}
