package controller

import (
	"context"
	"net/http"
	"slices"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-github/v56/github"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/k8s"
	"github.com/arikkfir/devbot/internal/util/lang"
)

var (
	RepositoryFinalizer = "repository.finalizers." + v1.GroupVersion.Group
)

type RepositoryReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	GitHubWebhookURL string
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

func (r *RepositoryReconciler) executeReconciliation(ctx context.Context, req controllerruntime.Request) *k8s.Result {
	// TODO: set finalizer that removes our webhook
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &v1.Repository{}, RepositoryFinalizer, nil)
	if result != nil {
		return result
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result
	}

	// Parse refresh interval
	var refreshInterval time.Duration
	if interval, result := r.parseRefreshInterval(rec); result != nil {
		return result
	} else {
		refreshInterval = interval
	}

	// Repository type-specific sections
	if rec.Object.Spec.GitHub != nil {
		return r.reconcileGitHubRepository(rec, refreshInterval)
	}

	// Unknown repository type
	rec.Object.Status.SetInvalidDueToUnknownRepositoryType("Unknown repository type")
	rec.Object.Status.SetUnauthenticatedDueToInvalid(rec.Object.Status.GetInvalidMessage())
	rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
	return rec.UpdateStatus()
}

func (r *RepositoryReconciler) parseRefreshInterval(rec *k8s.Reconciliation[*v1.Repository]) (time.Duration, *k8s.Result) {
	if interval, err := lang.ParseDuration(v1.MinRepositoryRefreshInterval, rec.Object.Spec.RefreshInterval); err != nil {
		rec.Object.Status.SetInvalidDueToInvalidRefreshInterval(err.Error())
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return 0, result
		}
		return 0, k8s.DoNotRequeue()
	} else {
		rec.Object.Status.SetValidIfInvalidDueToAnyOf(v1.InvalidRefreshInterval)
		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(v1.Invalid)
		return interval, rec.UpdateStatus()
	}
}

func (r *RepositoryReconciler) reconcileGitHubRepository(rec *k8s.Reconciliation[*v1.Repository], refreshInterval time.Duration) *k8s.Result {
	status := &rec.Object.Status

	// Update resolved-name if necessary
	resolvedName := rec.Object.Spec.GitHub.Owner + "/" + rec.Object.Spec.GitHub.Name
	if resolvedName != status.ResolvedName {
		status.ResolvedName = rec.Object.Spec.GitHub.Owner + "/" + rec.Object.Spec.GitHub.Name
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Connect to GitHub
	var ghc *github.Client
	if gitHubClient, result := r.connectToGitHub(rec, refreshInterval); result != nil {
		return result
	} else {
		ghc = gitHubClient
	}

	// Fetch the repository
	var ghRepo *github.Repository
	if ghr, result := r.fetchRepository(rec, refreshInterval, ghc); result != nil {
		return result
	} else {
		ghRepo = ghr
	}

	// Sync default branch
	if ghRepo.GetDefaultBranch() != rec.Object.Status.DefaultBranch {
		rec.Object.Status.DefaultBranch = ghRepo.GetDefaultBranch()
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Create missing ref objects based on current branches in the repository
	branchesToRevisionsMap := make(map[string]string)
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := ghc.Repositories.ListBranches(rec.Ctx, rec.Object.Spec.GitHub.Owner, rec.Object.Spec.GitHub.Name, branchesListOptions)
		if err != nil {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed listing branches: %+v", err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		}
		for _, branch := range branchesList {
			branchName := branch.GetName()
			revision := branch.GetCommit().GetSHA()
			branchesToRevisionsMap[branchName] = revision
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}
	status.Revisions = branchesToRevisionsMap
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(v1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Ensure webhook installed
	if result := r.ensureWebhook(rec, refreshInterval, ghc, ghRepo); result != nil {
		return result
	}

	// Done
	return k8s.RequeueAfter(refreshInterval)
}

func (r *RepositoryReconciler) connectToGitHub(rec *k8s.Reconciliation[*v1.Repository], refreshInterval time.Duration) (*github.Client, *k8s.Result) {
	status := &rec.Object.Status
	patCfg := rec.Object.Spec.GitHub.PersonalAccessToken

	// The GitHub client, to be initialized based on the authentication configuration selected
	var ghc *github.Client

	// Fetch secret
	authSecret := &v12.Secret{}
	secretObjKey := patCfg.Secret.GetObjectKey(rec.Object.Namespace)
	if err := r.Client.Get(rec.Ctx, secretObjKey, authSecret); err != nil {
		if errors.IsNotFound(err) {
			status.SetUnauthenticatedDueToAuthSecretNotFound("Secret '%s' not found", secretObjKey)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.RequeueAfter(refreshInterval)
		} else if errors.IsForbidden(err) {
			status.SetUnauthenticatedDueToAuthSecretForbidden("Secret '%s' is not accessible: %+v", secretObjKey, err)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.RequeueAfter(refreshInterval)
		} else {
			status.SetUnauthenticatedDueToInternalError("Failed reading secret '%s': %+v", secretObjKey, err)
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.RequeueAfter(refreshInterval)
		}
	}

	// Revert status if auth secret fetched successfully
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(v1.AuthSecretNotFound, v1.AuthSecretForbidden, v1.InternalError)
	status.SetCurrentIfStaleDueToAnyOf(v1.Unauthenticated)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	// Extract & validate personal access token
	pat, ok := authSecret.Data[patCfg.Key]
	if !ok {
		status.SetUnauthenticatedDueToAuthSecretKeyNotFound("Key '%s' not found in secret '%s'", patCfg.Key, secretObjKey)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}
		return nil, k8s.RequeueAfter(refreshInterval)
	} else if string(pat) == "" {
		status.SetUnauthenticatedDueToAuthTokenEmpty("Token in key '%s' in secret '%s' is empty", patCfg.Key, secretObjKey)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}
		return nil, k8s.RequeueAfter(refreshInterval)
	}

	// Revert status if auth secret fetched successfully
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(v1.AuthSecretKeyNotFound, v1.AuthTokenEmpty)
	status.SetCurrentIfStaleDueToAnyOf(v1.Unauthenticated)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	// Create the GitHub client & verify it's properly authenticated
	ghc = github.NewClient(nil).WithAuthToken(string(pat))
	if req, err := ghc.NewRequest("GET", "user", nil); err != nil {
		status.SetUnauthenticatedDueToAuthenticationFailed("Validation request creation failed: %+v", err)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}
		return nil, k8s.RequeueAfter(refreshInterval)
	} else if _, err := ghc.Do(rec.Ctx, req, nil); err != nil {
		status.SetUnauthenticatedDueToAuthenticationFailed("Validation request failed: %+v", err)
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}
		return nil, k8s.RequeueAfter(refreshInterval)
	}

	// Revert status if GitHub client is authenticated
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(v1.AuthenticationFailed)
	status.SetCurrentIfStaleDueToAnyOf(v1.Unauthenticated)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	return ghc, k8s.Continue()
}

func (r *RepositoryReconciler) fetchRepository(rec *k8s.Reconciliation[*v1.Repository], refreshInterval time.Duration, ghc *github.Client) (*github.Repository, *k8s.Result) {
	owner := rec.Object.Spec.GitHub.Owner
	name := rec.Object.Spec.GitHub.Name

	ghr, resp, err := ghc.Repositories.Get(rec.Ctx, owner, name)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rec.Object.Status.SetStaleDueToRepositoryNotFound("Repository not found: %s", resp.Status)
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.RequeueAfter(refreshInterval)
		} else {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed fetching repository '%s/%s': %+v", owner, name, err)
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.RequeueAfter(refreshInterval)
		}
	}

	// Revert status if set due to repository not found or internal error
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(v1.RepositoryNotFound, v1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	return ghr, nil
}

func (r *RepositoryReconciler) ensureWebhook(rec *k8s.Reconciliation[*v1.Repository], refreshInterval time.Duration, ghc *github.Client, repo *github.Repository) *k8s.Result {
	status := &rec.Object.Status

	if webhookCfg := rec.Object.Spec.GitHub.WebhookSecret; webhookCfg != nil {
		if r.GitHubWebhookURL == "" {
			status.SetInvalidDueToWebhooksNotEnabled("Webhooks not enabled - must provide webhooks URL to controller")
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.DoNotRequeue()
		}

		// Validate auth secret name & key
		if webhookCfg.Secret.Name == "" {
			status.SetInvalidDueToWebhookSecretNameMissing("Webhook secret name is empty")
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.DoNotRequeue()
		} else if webhookCfg.Key == "" {
			status.SetInvalidDueToWebhookSecretKeyMissing("Webhook secret key is missing")
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.DoNotRequeue()
		}

		// Revert invalid status if auth secret name & key are valid
		status.SetValidIfInvalidDueToAnyOf(v1.WebhookSecretKeyMissing, v1.WebhookSecretNameMissing)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		// Fetch secret
		webhookSecret := &v12.Secret{}
		secretObjKey := webhookCfg.Secret.GetObjectKey(rec.Object.Namespace)
		if err := r.Client.Get(rec.Ctx, secretObjKey, webhookSecret); err != nil {
			if errors.IsNotFound(err) {
				status.SetInvalidDueToWebhookSecretNotFound("Secret '%s' not found", secretObjKey)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.RequeueAfter(refreshInterval)
			} else if errors.IsForbidden(err) {
				status.SetInvalidDueToWebhookSecretForbidden("Secret '%s' is not accessible: %+v", secretObjKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.RequeueAfter(refreshInterval)
			} else {
				status.SetInvalidDueToInternalError("Failed reading secret '%s': %+v", secretObjKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.RequeueAfter(refreshInterval)
			}
		}

		// Revert status if auth secret fetched successfully
		status.SetValidIfInvalidDueToAnyOf(v1.WebhookSecretNotFound, v1.WebhookSecretForbidden, v1.InternalError)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		// Extract & validate personal access token
		secretValue, ok := webhookSecret.Data[webhookCfg.Key]
		if !ok {
			status.SetInvalidDueToWebhookSecretKeyNotFound("Key '%s' not found in secret '%s'", webhookCfg.Key, secretObjKey)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		} else if string(secretValue) == "" {
			status.SetInvalidDueToWebhookSecretEmpty("Key '%s' in secret '%s' is empty", webhookCfg.Key, secretObjKey)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(refreshInterval)
		}

		// Revert status if auth secret fetched successfully
		status.SetValidIfInvalidDueToAnyOf(v1.WebhookSecretKeyNotFound, v1.WebhookSecretEmpty)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		// Search for our webhook in the repository
		opts := &github.ListOptions{PerPage: 50}
		var webhook *github.Hook
		for webhook == nil {
			hooks, resp, err := ghc.Repositories.ListHooks(rec.Ctx, *repo.Owner.Login, *repo.Name, opts)
			if err != nil {
				status.SetInvalidDueToInternalError("Failed to list repository webhooks: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			}
			for _, hook := range hooks {
				if webhookURL, ok := hook.Config["url"]; ok {
					if webhookURL == r.GitHubWebhookURL {
						webhook = hook
						break
					}
				}
			}
			if resp.NextPage == 0 {
				break
			} else {
				opts.Page = resp.NextPage
			}
		}

		// If webhook not found, create it
		if webhook == nil {
			webhook = &github.Hook{
				Name: lang.Ptr("web"),
				Config: map[string]any{
					"url":          r.GitHubWebhookURL,
					"content_type": "json",
					"secret":       secretValue,
				},
				Events: []string{"push"},
				Active: lang.Ptr(true),
			}
			if _, _, err := ghc.Repositories.CreateHook(rec.Ctx, *repo.Owner.Login, *repo.Name, webhook); err != nil {
				status.SetInvalidDueToInternalError("Failed to create webhook: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
			}
			return k8s.Requeue()
		} else if contentType, ok := webhook.Config["content_type"]; !ok || contentType != "json" {
			config := &github.HookConfig{ContentType: lang.Ptr("json")}
			if _, _, err := ghc.Repositories.EditHookConfiguration(rec.Ctx, *repo.Owner.Login, *repo.Name, *webhook.ID, config); err != nil {
				status.SetInvalidDueToInternalError("Failed to update webhook config: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
			}
			return k8s.Requeue()
		} else if !slices.Contains(webhook.Events, "push") {
			webhook = &github.Hook{
				Config: map[string]any{
					"url":          r.GitHubWebhookURL,
					"content_type": "json",
					"secret":       secretValue,
				},
				Events: []string{"push"},
			}
			if _, _, err := ghc.Repositories.EditHook(rec.Ctx, *repo.Owner.Login, *repo.Name, *webhook.ID, webhook); err != nil {
				status.SetInvalidDueToInternalError("Failed to update webhook config: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
			}
			return k8s.Requeue()
		} else if status.LastWebhookPing == nil || time.Since(status.LastWebhookPing.Time) > 5*time.Minute {
			status.LastWebhookPing = lang.Ptr(metav1.NewTime(time.Now()))
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			if _, err := ghc.Repositories.PingHook(rec.Ctx, *repo.Owner.Login, *repo.Name, *webhook.ID); err != nil {
				status.SetInvalidDueToInternalError("Failed to ping webhook: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
			}
			return k8s.Requeue()
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).
		For(&v1.Repository{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Only reconcile if the generation has changed
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		})).
		Complete(r)
}
