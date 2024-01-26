package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func NewFetchGitHubBranchesAction(refreshInterval time.Duration, gh *github.Client, ghBranches *[]*github.Branch) reconcile.Action {
	return func(ctx context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		var branches []*github.Branch
		branchesListOptions := &github.BranchListOptions{}
		for {
			branchesList, response, err := gh.Repositories.ListBranches(ctx, r.Spec.Owner, r.Spec.Name, branchesListOptions)
			if err != nil {
				r.Status.SetMaybeStaleDueToGitHubAPIFailed("Failed listing branches: %+v", err)
				return reconcile.RequeueAfter(refreshInterval)
			}
			for _, branch := range branchesList {
				branches = append(branches, branch)
			}
			if response.NextPage == 0 {
				break
			}
			branchesListOptions.Page = response.NextPage
		}
		*ghBranches = branches
		return reconcile.Continue()
	}
}
