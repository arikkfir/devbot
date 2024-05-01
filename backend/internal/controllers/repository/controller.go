package repository

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/config"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"github.com/google/go-github/v56/github"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

var (
	Finalizer = "repository.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Config config.CommandConfig
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

func (r *Reconciler) executeReconciliation(ctx context.Context, req ctrl.Request) *k8s.Result {
	rec, result := k8s.NewReconciliation(ctx, r.Config, r.Client, req, &apiv1.Repository{}, Finalizer, nil)
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

func (r *Reconciler) parseRefreshInterval(rec *k8s.Reconciliation[*apiv1.Repository]) (time.Duration, *k8s.Result) {
	if interval, err := lang.ParseDuration(apiv1.MinRepositoryRefreshInterval, rec.Object.Spec.RefreshInterval); err != nil {
		rec.Object.Status.SetInvalidDueToInvalidRefreshInterval(err.Error())
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return 0, result
		}
		return 0, k8s.DoNotRequeue()
	} else {
		rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.InvalidRefreshInterval)
		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid)
		return interval, rec.UpdateStatus()
	}
}

func (r *Reconciler) reconcileGitHubRepository(rec *k8s.Reconciliation[*apiv1.Repository], refreshInterval time.Duration) *k8s.Result {
	status := &rec.Object.Status

	// Validate GitHub owner & name
	if rec.Object.Spec.GitHub.Owner == "" {
		status.ResolvedName = ""
		status.SetInvalidDueToRepositoryOwnerMissing("Repository owner is empty")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	} else if rec.Object.Spec.GitHub.Name == "" {
		status.ResolvedName = ""
		status.SetInvalidDueToRepositoryNameMissing("Repository name is empty")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	// Revert invalid status if owner & name are valid
	status.ResolvedName = rec.Object.Spec.GitHub.Owner + "/" + rec.Object.Spec.GitHub.Name
	status.SetValidIfInvalidDueToAnyOf(apiv1.RepositoryNameMissing, apiv1.RepositoryOwnerMissing)
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.Invalid)
	status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated)
	if result := rec.UpdateStatus(); result != nil {
		return result
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
		rec.Object.Status.SetStaleDueToDefaultBranchOutOfSync("Default branch is set to '%s' but should be '%s'", rec.Object.Status.DefaultBranch, ghRepo.GetDefaultBranch())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		rec.Object.Status.DefaultBranch = ghRepo.GetDefaultBranch()
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.DefaultBranchOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Create missing ref objects based on current branches in the repository
	branchesToRevisionsMap := make(map[string]string)
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := ghc.Repositories.ListBranches(rec.Ctx, rec.Object.Spec.GitHub.Owner, rec.Object.Spec.GitHub.Name, branchesListOptions)
		if err != nil {
			rec.Object.Status.SetMaybeStaleDueToBranchesOutOfSync("Failed listing branches: %+v", err)
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
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.BranchesOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Done
	return k8s.RequeueAfter(refreshInterval)
}

func (r *Reconciler) connectToGitHub(rec *k8s.Reconciliation[*apiv1.Repository], refreshInterval time.Duration) (*github.Client, *k8s.Result) {
	status := &rec.Object.Status

	// The GitHub client, to be initialized based on the authentication configuration selected
	var ghc *github.Client

	// If PAT authentication selected
	if patCfg := rec.Object.Spec.GitHub.PersonalAccessToken; patCfg != nil {

		// Validate auth secret name & key
		if patCfg.Secret.Name == "" {
			status.SetInvalidDueToAuthSecretNameMissing("Auth secret name is empty")
			status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.DoNotRequeue()
		} else if patCfg.Key == "" {
			status.SetInvalidDueToAuthSecretKeyMissing("Auth secret key is missing")
			status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
			status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
			if result := rec.UpdateStatus(); result != nil {
				return nil, result
			}
			return nil, k8s.DoNotRequeue()
		}

		// Revert invalid status if auth secret name & key are valid
		status.SetValidIfInvalidDueToAnyOf(apiv1.AuthSecretKeyMissing, apiv1.AuthSecretNameMissing)
		status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.Invalid)
		status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated)
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}

		// Fetch secret
		authSecret := &corev1.Secret{}
		secretObjKey := patCfg.Secret.GetObjectKey(rec.Object.Namespace)
		if err := r.Client.Get(rec.Ctx, secretObjKey, authSecret); err != nil {
			if apierrors.IsNotFound(err) {
				status.SetUnauthenticatedDueToAuthSecretNotFound("Secret '%s' not found", secretObjKey)
				status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
				if result := rec.UpdateStatus(); result != nil {
					return nil, result
				}
				return nil, k8s.RequeueAfter(refreshInterval)
			} else if apierrors.IsForbidden(err) {
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
		status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.AuthSecretNotFound, apiv1.AuthSecretForbidden, apiv1.InternalError)
		status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated)
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
		status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.AuthSecretKeyNotFound, apiv1.AuthTokenEmpty)
		status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated)
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}

		ghc = github.NewClient(nil).WithAuthToken(string(pat))

	} else {
		// No authentication selected

		status.SetInvalidDueToAuthConfigMissing("Auth config is missing")
		status.SetUnauthenticatedDueToInvalid(status.GetInvalidMessage())
		status.SetMaybeStaleDueToUnauthenticated(status.GetUnauthenticatedMessage())
		if result := rec.UpdateStatus(); result != nil {
			return nil, result
		}
		return nil, k8s.DoNotRequeue()
	}

	// Revert invalid status if set due to missing auth configuration
	status.SetValidIfInvalidDueToAnyOf(apiv1.AuthConfigMissing)
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.Invalid)
	status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	// Verify the GitHub client is authenticated
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
	status.SetAuthenticatedIfUnauthenticatedDueToAnyOf(apiv1.AuthenticationFailed)
	status.SetCurrentIfStaleDueToAnyOf(apiv1.Unauthenticated)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	return ghc, k8s.Continue()
}

func (r *Reconciler) fetchRepository(rec *k8s.Reconciliation[*apiv1.Repository], refreshInterval time.Duration, ghc *github.Client) (*github.Repository, *k8s.Result) {
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
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return nil, result
	}

	return ghr, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Repository{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Only reconcile if the generation has changed
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		})).
		Complete(r)
}
