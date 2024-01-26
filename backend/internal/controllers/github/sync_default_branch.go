package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSyncDefaultBranchAction(ghRepo *github.Repository) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		if ghRepo.GetDefaultBranch() != r.Status.DefaultBranch {
			r.Status.SetStaleDueToDefaultBranchOutOfSync("Default branch is out of sync (current: '%s', desired: '%s')", r.Status.DefaultBranch, ghRepo.GetDefaultBranch())
			r.Status.DefaultBranch = ghRepo.GetDefaultBranch()
			return reconcile.Requeue()
		}
		return reconcile.Continue()
	}
}
