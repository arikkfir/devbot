package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/reconciler"
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
	rec, result := reconciler.NewReconciliation(ctx, r.Client, req, &apiv1.GitHubRepositoryRef{}, RefFinalizer, nil)
	if result != nil {
		return result.Return()
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result.Return()
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result.Return()
	}

	// Get controlling repository
	repo := &apiv1.GitHubRepository{}
	if result := rec.GetRequiredController(repo); result != nil {
		return result.Return()
	}

	// Parse refresh interval
	var refreshInterval time.Duration
	if result := rec.ParseRefreshInterval(repo.Spec.RefreshInterval, &refreshInterval, rec.Object.Status.SetInvalidDueToInvalidRefreshInterval, rec.Object.Status.SetMaybeStaleDueToInvalid); result != nil {
		return result.Return()
	} else if rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.InvalidRefreshInterval) || rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid) {
		if result := rec.UpdateStatus(reconciler.WithStrategy(reconciler.Continue)); result != nil {
			return result.Return()
		}
	}

	//Sync ref repository owner
	if repo.Spec.Owner != rec.Object.Status.RepositoryOwner {
		rec.Object.Status.SetStaleDueToRepositoryOwnerOutOfSync("Repository owner '%s' is stale (expected '%s')", rec.Object.Status.RepositoryOwner, repo.Spec.Owner)
		rec.Object.Status.RepositoryOwner = repo.Spec.Owner
		return rec.UpdateStatus(reconciler.WithStrategy(reconciler.Requeue)).Return()
	} else if rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryOwnerOutOfSync) {
		if result := rec.UpdateStatus(reconciler.WithStrategy(reconciler.Continue)); result != nil {
			return result.Return()
		}
	}

	// Sync ref repository name
	if repo.Spec.Name != rec.Object.Status.RepositoryName {
		rec.Object.Status.SetStaleDueToRepositoryNameOutOfSync("Repository name '%s' is stale (expected '%s')", rec.Object.Status.RepositoryName, repo.Spec.Name)
		rec.Object.Status.RepositoryName = repo.Spec.Name
		return rec.UpdateStatus(reconciler.WithStrategy(reconciler.Requeue)).Return()
	} else if rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNameOutOfSync) {
		if result := rec.UpdateStatus(reconciler.WithStrategy(reconciler.Continue)); result != nil {
			return result.Return()
		}
	}

	// Connect to GitHub & fetch repository
	var gh *github.Client
	var ghRepo *github.Repository
	if result := ConnectToGitHub(rec, repo.Spec.Owner, repo.Spec.Name, repo.Spec.Auth, refreshInterval, &gh, &ghRepo); result != nil {
		return result.Return()
	}

	// Fetch branch details from GitHub
	branch, response, err := gh.Repositories.GetBranch(ctx, rec.Object.Status.RepositoryOwner, rec.Object.Status.RepositoryName, rec.Object.Spec.Ref, 0)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			if err := r.Delete(ctx, rec.Object); err != nil {
				if apierrors.IsNotFound(err) {
					return reconciler.DoNotRequeue().Return()
				} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
					return reconciler.Requeue().Return()
				} else {
					return reconciler.RequeueDueToError(err).Return()
				}
			}
			return reconciler.DoNotRequeue().Return()
		} else {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed fetching branch '%s' from GitHub: %+v", rec.Object.Spec.Ref, err)
			return rec.UpdateStatus(reconciler.WithStrategy(reconciler.Requeue)).Return()
		}
	}

	// Sync commit SHA
	commitSHA := branch.GetCommit().GetSHA()
	if commitSHA != rec.Object.Status.CommitSHA {
		rec.Object.Status.SetStaleDueToCommitSHAOutOfSync("Commit SHA '%s' is stale (expected '%s')", rec.Object.Status.CommitSHA, commitSHA)
		rec.Object.Status.CommitSHA = commitSHA
		return rec.UpdateStatus(reconciler.WithStrategy(reconciler.Requeue)).Return()
	} else if rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.CommitSHAOutOfSync) {
		if result := rec.UpdateStatus(reconciler.WithStrategy(reconciler.Continue)); result != nil {
			return result.Return()
		}
	}

	// Done
	return reconciler.RequeueAfter(refreshInterval).Return()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Watches(&apiv1.GitHubRepository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			repo := o.(*apiv1.GitHubRepository)
			if repo.APIVersion == "" {
				panic("APIVersion is empty")
			} else if repo.Kind == "" {
				panic("Kind is empty")
			}

			refsList := &apiv1.GitHubRepositoryRefList{}
			if err := r.List(ctx, refsList, reconciler.OwnedBy(r.Scheme, repo)); err != nil {
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
