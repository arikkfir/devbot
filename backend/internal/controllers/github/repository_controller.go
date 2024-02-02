package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
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

func (r *RepositoryReconciler) CreateMissingRefObjects(rec *k8s.Reconciliation[*apiv1.GitHubRepository], gh *github.Client, refsList *apiv1.GitHubRepositoryRefList) *k8s.Result {
	o := rec.Object

	var branchesWithoutRefs []string
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := gh.Repositories.ListBranches(rec.Ctx, o.Spec.Owner, o.Spec.Name, branchesListOptions)
		if err != nil {
			o.Status.SetMaybeStaleDueToGitHubAPIFailure("Failed listing branches: %+v", err)
			return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
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
				o.Status.SetStaleDueToBranchesOutOfSync("Branches without GitHubRepositoryRef object found: %s", strings.Join(branchesWithoutRefs, ", "))
				if result := rec.UpdateStatus(k8s.WithStrategy(k8s.Continue)); result != nil {
					return result
				}

				ref := &apiv1.GitHubRepositoryRef{
					ObjectMeta: metav1.ObjectMeta{
						Name:            stringsutil.RandomHash(7),
						Namespace:       o.Namespace,
						OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(o, apiv1.GitHubRepositoryGVK)},
					},
					Spec: apiv1.GitHubRepositoryRefSpec{Ref: branch.GetName()},
				}
				if err := r.Client.Create(rec.Ctx, ref); err != nil {
					o.Status.SetStaleDueToInternalError("Failed creating ref object for branch '%s': %+v", branch.GetName(), err)
					return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
				}
			}
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}

	if o.Status.SetCurrentIfStaleDueToAnyOf(apiv1.GitHubAPIFailure, apiv1.BranchesOutOfSync, apiv1.InternalError) {
		return rec.UpdateStatus(k8s.WithStrategy(k8s.Continue))
	} else {
		return k8s.Continue()
	}
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.GitHubRepository{}, RepositoryFinalizer, nil)
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

	// Parse refresh interval
	var refreshInterval time.Duration
	if result := rec.ParseRefreshInterval(rec.Object.Spec.RefreshInterval, &refreshInterval, rec.Object.Status.SetInvalidDueToInvalidRefreshInterval, rec.Object.Status.SetMaybeStaleDueToInvalid); result != nil {
		return result.Return()
	} else if rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.InvalidRefreshInterval) || rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid) {
		if result := rec.UpdateStatus(k8s.WithStrategy(k8s.Continue)); result != nil {
			return result.Return()
		}
	}

	// Connect to GitHub & fetch repository
	var gh *github.Client
	var ghRepo *github.Repository
	if result := ConnectToGitHub(rec, rec.Object.Spec.Owner, rec.Object.Spec.Name, rec.Object.Spec.Auth, refreshInterval, &gh, &ghRepo); result != nil {
		return result.Return()
	}

	// Sync default branch
	if ghRepo.GetDefaultBranch() != rec.Object.Status.DefaultBranch {
		rec.Object.Status.SetStaleDueToDefaultBranchOutOfSync("Default branch is set to '%s' but should be '%s'", rec.Object.Status.DefaultBranch, ghRepo.GetDefaultBranch())
		rec.Object.Status.DefaultBranch = ghRepo.GetDefaultBranch()
		return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue)).Return()
	} else if rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.DefaultBranchOutOfSync) {
		if result := rec.UpdateStatus(k8s.WithStrategy(k8s.Continue)); result != nil {
			return result.Return()
		}
	}

	// Fetch existing ref objects
	refsList := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, refsList, k8s.OwnedBy(r.Client.Scheme(), rec.Object)); err != nil {
		return k8s.RequeueDueToError(errors.New("failed listing owned objects: %w", err)).Return()
	}

	// Create missing ref objects based on current branches in the repository
	if result := r.CreateMissingRefObjects(rec, gh, refsList); result != nil {
		return result.Return()
	}

	// Done
	return k8s.RequeueAfter(refreshInterval).Return()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Not watching GitHubRepositoryRef because nothing in repository objects is affected by their state
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
