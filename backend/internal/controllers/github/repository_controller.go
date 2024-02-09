package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

var (
	RepositoryFinalizer = "repository.github.finalizers." + apiv1.GroupVersion.Group
)

type RepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.GitHubRepository{}, RepositoryFinalizer, nil)
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

	// Parse refresh interval
	var refreshInterval time.Duration
	if interval, err := lang.ParseDuration(apiv1.MinGitHubRepositoryRefreshInterval, rec.Object.Spec.RefreshInterval); err != nil {
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

	// Connect to GitHub & fetch repository
	var gh *github.Client
	var ghRepo *github.Repository
	if result := ConnectToGitHub(rec, rec.Object.Spec.Owner, rec.Object.Spec.Name, rec.Object.Spec.Auth, refreshInterval, &gh, &ghRepo); result != nil {
		return result.ToResultAndError()
	}

	// Sync default branch
	if ghRepo.GetDefaultBranch() != rec.Object.Status.DefaultBranch {
		rec.Object.Status.SetStaleDueToDefaultBranchOutOfSync("Default branch is set to '%s' but should be '%s'", rec.Object.Status.DefaultBranch, ghRepo.GetDefaultBranch())
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		rec.Object.Status.DefaultBranch = ghRepo.GetDefaultBranch()
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.DefaultBranchOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Fetch existing ref objects
	refsList := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, refsList, k8s.OwnedBy(r.Client.Scheme(), rec.Object)); err != nil {
		return k8s.RequeueDueToError(errors.New("failed listing owned objects: %w", err)).ToResultAndError()
	}

	// Create missing ref objects based on current branches in the repository
	var branchesWithoutRefs []string
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := gh.Repositories.ListBranches(rec.Ctx, rec.Object.Spec.Owner, rec.Object.Spec.Name, branchesListOptions)
		if err != nil {
			rec.Object.Status.SetMaybeStaleDueToGitHubAPIFailure("Failed listing branches: %+v", err)
			if result := rec.UpdateStatus(); result != nil {
				return result.ToResultAndError()
			}
			return k8s.RequeueAfter(refreshInterval).ToResultAndError()
		}
		for _, branch := range branchesList {

			found := false
			for _, ref := range refsList.Items {
				if ref.Spec.Ref == branch.GetName() {
					found = true
					break
				}
			}
			if !found {
				branchesWithoutRefs = append(branchesWithoutRefs, branch.GetName())

				rec.Object.Status.SetStaleDueToBranchesOutOfSync("Branches without GitHubRepositoryRef object found: %s", strings.Join(branchesWithoutRefs, ", "))
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}

				ref := &apiv1.GitHubRepositoryRef{
					ObjectMeta: metav1.ObjectMeta{
						Name:            stringsutil.RandomHash(7),
						Namespace:       rec.Object.Namespace,
						OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.GitHubRepositoryGVK)},
					},
					Spec: apiv1.GitHubRepositoryRefSpec{Ref: branch.GetName()},
				}
				if err := r.Client.Create(rec.Ctx, ref); err != nil {
					rec.Object.Status.SetStaleDueToInternalError("Failed creating ref object for branch '%s': %+v", branch.GetName(), err)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				}
			}
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.GitHubAPIFailure, apiv1.BranchesOutOfSync, apiv1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Done
	return k8s.DoNotRequeue().ToResultAndError()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Not watching GitHubRepositoryRef because nothing in repository objects is affected by their state
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
