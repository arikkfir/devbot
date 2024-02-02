package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"github.com/google/go-github/v56/github"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	RefFinalizer = "ref." + RepositoryFinalizer
)

type RefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.GitHubRepositoryRef{}, RefFinalizer, nil)
	if result != nil {
		return result.ToResultAndError()
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result.ToResultAndError()
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result.ToResultAndError()
	}

	// Get controlling repository
	repo := &apiv1.GitHubRepository{}
	if result := rec.GetRequiredController(repo); result != nil {
		return result.ToResultAndError()
	}

	// Parse refresh interval
	var refreshInterval time.Duration
	if interval, err := lang.ParseDuration(apiv1.MinimumRefreshIntervalSeconds, repo.Spec.RefreshInterval); err != nil {
		rec.Object.Status.SetInvalidDueToInvalidRefreshInterval(err.Error())
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		return k8s.DoNotRequeue().ToResultAndError()
	} else {
		rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.InvalidRefreshInterval)
		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid)
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		refreshInterval = interval
	}

	//Sync ref repository owner
	if repo.Spec.Owner != rec.Object.Status.RepositoryOwner {
		rec.Object.Status.SetStaleDueToRepositoryOwnerOutOfSync("Repository owner '%s' is stale (expected '%s')", rec.Object.Status.RepositoryOwner, repo.Spec.Owner)
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		rec.Object.Status.RepositoryOwner = repo.Spec.Owner
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryOwnerOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Sync ref repository name
	if repo.Spec.Name != rec.Object.Status.RepositoryName {
		rec.Object.Status.SetStaleDueToRepositoryNameOutOfSync("Repository name '%s' is stale (expected '%s')", rec.Object.Status.RepositoryName, repo.Spec.Name)
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		rec.Object.Status.RepositoryName = repo.Spec.Name
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNameOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Connect to GitHub & fetch repository
	var gh *github.Client
	var ghRepo *github.Repository
	if result := ConnectToGitHub(rec, repo.Spec.Owner, repo.Spec.Name, repo.Spec.Auth, refreshInterval, &gh, &ghRepo); result != nil {
		return result.ToResultAndError()
	}

	// Fetch branch details from GitHub
	branch, response, err := gh.Repositories.GetBranch(ctx, rec.Object.Status.RepositoryOwner, rec.Object.Status.RepositoryName, rec.Object.Spec.Ref, 0)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			if err := r.Delete(ctx, rec.Object); err != nil {
				if apierrors.IsNotFound(err) {
					return k8s.DoNotRequeue().ToResultAndError()
				} else if apierrors.IsConflict(err) {
					return k8s.Requeue().ToResultAndError()
				} else {
					return k8s.RequeueDueToError(err).ToResultAndError()
				}
			}
			return k8s.DoNotRequeue().ToResultAndError()
		} else {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed fetching branch '%s' from GitHub: %+v", rec.Object.Spec.Ref, err)
			if result := rec.UpdateStatus(); result != nil {
				return result.ToResultAndError()
			}
			return k8s.Requeue().ToResultAndError()
		}
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Sync commit SHA
	commitSHA := branch.GetCommit().GetSHA()
	if commitSHA != rec.Object.Status.CommitSHA {
		rec.Object.Status.SetStaleDueToCommitSHAOutOfSync("Commit SHA '%s' is stale (expected '%s')", rec.Object.Status.CommitSHA, commitSHA)
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		rec.Object.Status.CommitSHA = commitSHA
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.CommitSHAOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Done
	return k8s.RequeueAfter(refreshInterval).ToResultAndError()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Watches(&apiv1.GitHubRepository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			repo := o.(*apiv1.GitHubRepository)

			refsList := &apiv1.GitHubRepositoryRefList{}
			if err := r.List(ctx, refsList, k8s.OwnedBy(r.Scheme, repo)); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list GitHubRepositoryRef objects")
				return nil
			}

			var requests []reconcile.Request
			for _, ref := range refsList.Items {
				requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKey{Namespace: ref.Namespace, Name: ref.Name}})
			}
			return requests
		})).
		Complete(r)
}
