package github

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/google/go-github/v56/github"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ConnectionStatus interface {
	GetInvalidMessage() string
	GetUnauthenticatedMessage() string
	SetAuthenticatedIfUnauthenticatedDueToAnyOf(...string) bool
	SetCurrentIfStaleDueToAnyOf(...string) bool
	SetInvalidDueToAuthConfigMissing(string, ...interface{}) bool
	SetInvalidDueToAuthSecretKeyMissing(string, ...interface{}) bool
	SetInvalidDueToAuthSecretNameMissing(string, ...interface{}) bool
	SetInvalidDueToRepositoryNameMissing(string, ...interface{}) bool
	SetInvalidDueToRepositoryOwnerMissing(string, ...interface{}) bool
	SetMaybeStaleDueToInternalError(string, ...interface{}) bool
	SetMaybeStaleDueToUnauthenticated(string, ...interface{}) bool
	SetStaleDueToRepositoryNotFound(string, ...interface{}) bool
	SetUnauthenticatedDueToAuthSecretForbidden(string, ...interface{}) bool
	SetUnauthenticatedDueToAuthSecretGetFailed(string, ...interface{}) bool
	SetUnauthenticatedDueToAuthSecretKeyNotFound(string, ...interface{}) bool
	SetUnauthenticatedDueToAuthSecretNotFound(string, ...interface{}) bool
	SetUnauthenticatedDueToAuthTokenEmpty(string, ...interface{}) bool
	SetUnauthenticatedDueToInvalid(string, ...interface{}) bool
	SetUnauthenticatedDueToTokenValidationFailed(string, ...interface{}) bool
	SetValidIfInvalidDueToAnyOf(...string) bool
}

func ConnectToGitHub[O client.Object](r *k8s.Reconciliation[O], repoOwner, repoName string, auth apiv1.GitHubRepositoryAuth, refreshInterval time.Duration, gh **github.Client, ghRepo **github.Repository) *k8s.Result {
	status := k8s.MustGetStatusOfType[ConnectionStatus](r.Object)
	patCfg := auth.PersonalAccessToken

	if patCfg == nil {
		status.SetInvalidDueToAuthConfigMissing("Auth config is missing")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	authSecretName := patCfg.Secret.Name
	if authSecretName == "" {
		status.SetInvalidDueToAuthSecretNameMissing("Auth secret name is empty")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	secretKey := patCfg.Key
	if secretKey == "" {
		status.SetInvalidDueToAuthSecretKeyMissing("Auth secret key is missing")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	if repoOwner == "" {
		status.SetInvalidDueToRepositoryOwnerMissing("Repository owner is empty")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	} else if repoName == "" {
		status.SetInvalidDueToRepositoryNameMissing("Repository name is empty")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	status.SetValidIfInvalidDueToAnyOf(apiv1.AuthConfigMissing, apiv1.AuthSecretNameMissing, apiv1.AuthSecretKeyMissing, apiv1.RepositoryOwnerMissing, apiv1.RepositoryNameMissing)
	if result := r.UpdateStatus(); result != nil {
		return result
	}

	authSecret := &corev1.Secret{}
	secretObjKey := patCfg.Secret.GetObjectKey(r.Object.GetNamespace())
	if err := r.Client.Get(r.Ctx, secretObjKey, authSecret); err != nil {
		if apierrors.IsNotFound(err) {
			status.SetUnauthenticatedDueToAuthSecretNotFound("Secret '%s' not found", secretObjKey)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		} else if apierrors.IsForbidden(err) {
			status.SetUnauthenticatedDueToAuthSecretForbidden("Secret '%s' is not accessible: %+v", secretObjKey, err)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		} else {
			status.SetUnauthenticatedDueToAuthSecretGetFailed("Failed reading secret '%s': %+v", secretObjKey, err)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		}
	}

	pat, ok := authSecret.Data[secretKey]
	if !ok {
		status.SetUnauthenticatedDueToAuthSecretKeyNotFound("Key '%s' not found in secret '%s'", secretKey, secretObjKey)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.RequeueAfter(refreshInterval)
	} else if string(pat) == "" {
		status.SetUnauthenticatedDueToAuthTokenEmpty("Token in key '%s' in secret '%s' is empty", secretKey, secretObjKey)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.RequeueAfter(refreshInterval)
	}

	ghc := github.NewClient(nil).WithAuthToken(string(pat))
	if req, err := ghc.NewRequest("GET", "user", nil); err != nil {
		status.SetUnauthenticatedDueToTokenValidationFailed("Token validation request creation failed: %+v", err)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.RequeueAfter(refreshInterval)
	} else if _, err := ghc.Do(r.Ctx, req, nil); err != nil {
		status.SetUnauthenticatedDueToTokenValidationFailed("Token validation request failed: %+v", err)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return k8s.RequeueAfter(refreshInterval)
	}

	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.Invalid, apiv1.AuthSecretNotFound, apiv1.AuthSecretForbidden, apiv1.AuthSecretGetFailed, apiv1.AuthSecretKeyNotFound, apiv1.AuthTokenEmpty, apiv1.TokenValidationFailed)
	if result := r.UpdateStatus(); result != nil {
		return result
	}

	ghr, resp, err := ghc.Repositories.Get(r.Ctx, repoOwner, repoName)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			status.SetStaleDueToRepositoryNotFound("Repository not found: %s", resp.Status)
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		} else {
			status.SetMaybeStaleDueToInternalError("Failed fetching repository '%s/%s': %+v", repoOwner, repoName, err)
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		}
	}

	status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated, apiv1.RepositoryNotFound, apiv1.InternalError)
	if result := r.UpdateStatus(); result != nil {
		return result
	}

	*gh = ghc
	*ghRepo = ghr
	return k8s.Continue()
}
