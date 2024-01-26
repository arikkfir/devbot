package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	Finalizer = "repository.github.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var refreshInterval time.Duration
	var gh *github.Client
	var ghRepo *github.Repository
	var ghBranches []*github.Branch
	var ghRefsList = &apiv1.GitHubRepositoryRefList{}
	reconciliation := reconcile.Reconciliation{
		Actions: []reconcile.Action{
			reconcile.NewFinalizeAction(Finalizer, nil),
			reconcile.NewAddFinalizerAction(Finalizer),
			reconcile.NewGetOwnedObjectsOfTypeAction(ghRefsList),
			NewParseRefreshInterval(&refreshInterval),
			NewConnectToGitHubAction(refreshInterval, &gh),
			NewFetchGitHubRepositoryAction(refreshInterval, gh, &ghRepo),
			NewFetchGitHubBranchesAction(refreshInterval, gh, &ghBranches),
			NewSyncDefaultBranchAction(ghRepo),
			NewCreateMissingGitHubRepositoryRefObjectsAction(ghBranches, ghRefsList),
			NewSyncGitHubRepositoryRefObjectsAction(ghBranches, ghRefsList),
			reconcile.NewRequeueAfterAction(refreshInterval),
		},
	}
	return reconciliation.Execute(ctx, r.Client, req, &apiv1.GitHubRepository{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Owns(&apiv1.GitHubRepositoryRef{}).
		Complete(r)
}
