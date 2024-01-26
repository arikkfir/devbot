package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func NewConnectToGitHubAction(refreshInterval time.Duration, gh **github.Client) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		auth := r.Spec.Auth
		patCfg := auth.PersonalAccessToken

		if patCfg == nil {
			r.Status.SetInvalidDueToAuthConfigMissing("Auth config is missing")
			r.Status.SetUnauthenticatedDueToInvalid(r.Status.GetInvalidMessage())
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.DoNotRequeue()
		}

		authSecretName := patCfg.Secret.Name
		if authSecretName == "" {
			r.Status.SetInvalidDueToAuthSecretNameMissing("Auth secret name is empty")
			r.Status.SetUnauthenticatedDueToInvalid(r.Status.GetInvalidMessage())
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.DoNotRequeue()
		}

		authSecret := &corev1.Secret{}
		secretObjKey := patCfg.Secret.GetObjectKey(o.GetNamespace())
		if err := c.Get(ctx, secretObjKey, authSecret); err != nil {
			if apierrors.IsNotFound(err) {
				r.Status.SetUnauthenticatedDueToAuthSecretNotFound("Secret '%s' not found", secretObjKey)
				r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
				return reconcile.RequeueAfter(refreshInterval)
			} else if apierrors.IsForbidden(err) {
				r.Status.SetUnauthenticatedDueToAuthSecretForbidden("Secret '%s' is not accessible: %+v", secretObjKey, err)
				r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
				return reconcile.RequeueAfter(refreshInterval)
			} else {
				r.Status.SetUnauthenticatedDueToAuthSecretGetFailed("Failed reading secret '%s': %+v", secretObjKey, err)
				r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
				return reconcile.RequeueDueToError(err)
			}
		}

		secretKey := patCfg.Key
		if secretKey == "" {
			r.Status.SetInvalidDueToAuthSecretKeyMissing("Auth secret key is missing")
			r.Status.SetUnauthenticatedDueToInvalid(r.Status.GetInvalidMessage())
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.DoNotRequeue()
		}

		pat, ok := authSecret.Data[secretKey]
		if !ok {
			r.Status.SetUnauthenticatedDueToAuthSecretKeyNotFound("Key '%s' not found in secret '%s'", secretKey, secretObjKey)
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.RequeueAfter(refreshInterval)
		} else if string(pat) == "" {
			r.Status.SetUnauthenticatedDueToAuthTokenEmpty("Token in key '%s' in secret '%s' is empty", secretKey, secretObjKey)
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.RequeueAfter(refreshInterval)
		}

		ghc := github.NewClient(nil).WithAuthToken(string(pat))
		if req, err := ghc.NewRequest("GET", "user", nil); err != nil {
			r.Status.SetUnauthenticatedDueToTokenValidationFailed("Token validation request creation failed: %+v", err)
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.RequeueAfter(refreshInterval)
		} else if _, err := ghc.Do(ctx, req, nil); err != nil {
			r.Status.SetUnauthenticatedDueToTokenValidationFailed("Token validation request failed: %+v", err)
			r.Status.SetMaybeStaleDueToUnauthenticated(r.Status.GetUnauthenticatedMessage())
			return reconcile.RequeueAfter(refreshInterval)
		}

		*gh = ghc
		return reconcile.Continue()
	}
}
